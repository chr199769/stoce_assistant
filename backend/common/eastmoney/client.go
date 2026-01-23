package eastmoney

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
		}

		clonedReq := req.Clone(req.Context())
		clonedReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		clonedReq.Header.Set("Referer", "https://quote.eastmoney.com/")
		clonedReq.Header.Set("Origin", "https://quote.eastmoney.com")
		clonedReq.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		clonedReq.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		clonedReq.Header.Set("Connection", "keep-alive")
		clonedReq.Header.Set("Sec-Fetch-Dest", "empty")
		clonedReq.Header.Set("Sec-Fetch-Mode", "cors")
		clonedReq.Header.Set("Sec-Fetch-Site", "same-site")

		resp, err := c.httpClient.Do(clonedReq)
		if err != nil {
			if attempt < 2 && isRetryableError(err) {
				lastErr = err
				continue
			}
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			err = fmt.Errorf("eastmoney api 错误: 状态码=%d 内容=%s", resp.StatusCode, string(body))
			if attempt < 2 && isRetryableStatus(resp.StatusCode) {
				lastErr = err
				continue
			}
			return nil, err
		}

		return resp, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("eastmoney api 请求失败")
}

func isRetryableError(err error) bool {
	if errors.Is(err, io.EOF) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}
	return false
}

func isRetryableStatus(status int) bool {
	if status == http.StatusTooManyRequests {
		return true
	}
	return status >= 500
}

type FinancialReportResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Data []struct {
			ReportDate      string  `json:"REPORTDATE"`
			TotalIncome     float64 `json:"TOTAL_OPERATE_INCOME"`
			ParentNetProfit float64 `json:"PARENT_NETPROFIT"`
			BasicEPS        float64 `json:"BASIC_EPS"`
			TotalIncomeYOY  float64 `json:"YSTZ"`
			NetProfitYOY    float64 `json:"SJLTZ"`
		} `json:"data"`
	} `json:"result"`
}

func (c *Client) GetFinancialReports(ctx context.Context, code string) ([]*FinancialData, error) {
	cleanCode := code
	if len(code) > 6 {
		cleanCode = code[len(code)-6:]
	}

	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_LICO_FN_CPD&columns=ALL&filter=(SECURITY_CODE=%%22%s%%22)&pageNumber=1&pageSize=5&sortTypes=-1&sortColumns=REPORTDATE", cleanCode)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result FinancialReportResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		if len(result.Result.Data) == 0 {
			return []*FinancialData{}, nil
		}
		return nil, fmt.Errorf("eastmoney api 失败")
	}

	var reports []*FinancialData
	for _, item := range result.Result.Data {
		date := item.ReportDate
		if len(date) > 10 {
			date = date[:10]
		}

		reports = append(reports, &FinancialData{
			ReportDate:   date,
			TotalRevenue: item.TotalIncome,
			NetProfit:    item.ParentNetProfit,
			Eps:          item.BasicEPS,
			RevenueYoy:   item.TotalIncomeYOY,
			ProfitYoy:    item.NetProfitYOY,
		})
	}

	return reports, nil
}

type SectorRankResponse struct {
	Rc   int `json:"rc"`
	Data *struct {
		Total int `json:"total"`
		Diff  []struct {
			Code           string  `json:"f12"`
			Name           string  `json:"f14"`
			ChangePercent  float64 `json:"f3"`
			NetInflow      float64 `json:"f62"`
			TopStockName   string  `json:"f128"`
			TopStockCode   string  `json:"f140"`
			TopStockChange float64 `json:"f136"`
		} `json:"diff"`
	} `json:"data"`
}

func (c *Client) GetSectorRank(ctx context.Context, rankType string, limit int) ([]*SectorInfo, error) {
	if limit <= 0 {
		limit = 20
	}

	var fs string
	if rankType == "industry" {
		fs = "m:90%2Bt:2%2Bf:!50"
	} else {
		fs = "m:90%2Bt:3%2Bf:!50"
	}

	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=%d&po=1&np=1&fltt=2&invt=2&fid=f3&fs=%s&fields=f12,f13,f14,f2,f3,f62,f128,f140,f136", limit, fs)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SectorRankResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Data == nil {
		return nil, fmt.Errorf("eastmoney 未返回数据")
	}

	var sectors []*SectorInfo
	for _, item := range result.Data.Diff {
		sectors = append(sectors, &SectorInfo{
			Code:          item.Code,
			Name:          item.Name,
			ChangePercent: item.ChangePercent,
			NetInflow:     item.NetInflow,
			TopStockName:  item.TopStockName,
			TopStockCode:  item.TopStockCode,
		})
	}

	return sectors, nil
}

type SectorStocksResponse struct {
	Rc   int `json:"rc"`
	Data *struct {
		Total int `json:"total"`
		Diff  []struct {
			Code          string  `json:"f12"`
			Name          string  `json:"f14"`
			Price         float64 `json:"f2"`
			ChangePercent float64 `json:"f3"`
			Volume        int64   `json:"f5"`
			Amount        float64 `json:"f6"`
			MarketCap     float64 `json:"f20"`
		} `json:"diff"`
	} `json:"data"`
}

func (c *Client) GetSectorStocksRaw(ctx context.Context, sectorCode string) ([]*SectorStockItem, error) {
	fs := fmt.Sprintf("b:%s", sectorCode)
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=100&po=1&np=1&fltt=2&invt=2&fid=f3&fs=%s&fields=f12,f14,f2,f3,f5,f6,f20", fs)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SectorStocksResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Data == nil {
		return nil, fmt.Errorf("eastmoney 未返回数据")
	}

	var stocks []*SectorStockItem
	for _, item := range result.Data.Diff {
		stocks = append(stocks, &SectorStockItem{
			Code:          item.Code,
			Name:          item.Name,
			Price:         item.Price,
			ChangePercent: item.ChangePercent,
			Volume:        item.Volume,
			Amount:        item.Amount,
			MarketCap:     item.MarketCap,
		})
	}
	return stocks, nil
}

type DragonTigerListResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Data []struct {
			SecurityCode    string  `json:"SECURITY_CODE"`
			SecurityName    string  `json:"SECURITY_NAME_ABBR"`
			ClosePrice      float64 `json:"CLOSE_PRICE"`
			ChangeRate      float64 `json:"CHANGE_RATE"`
			Explain         string  `json:"EXPLANATION"`
			BillBoardNetAmt float64 `json:"BILLBOARD_NET_AMT"`
		} `json:"data"`
	} `json:"result"`
}

func (c *Client) GetDragonTigerList(ctx context.Context, date string) ([]*DragonTigerItem, error) {
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_DAILYBILLBOARD_DETAILS&columns=ALL&filter=(TRADE_DATE=%%27%s%%27)&pageNumber=1&pageSize=100&sortTypes=-1&sortColumns=BILLBOARD_NET_AMT", date)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result DragonTigerListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return []*DragonTigerItem{}, nil
	}

	var items []*DragonTigerItem
	for _, d := range result.Result.Data {
		items = append(items, &DragonTigerItem{
			Date:          date,
			Code:          d.SecurityCode,
			Name:          d.SecurityName,
			ClosePrice:    d.ClosePrice,
			ChangePercent: d.ChangeRate,
			Reason:        d.Explain,
			NetInflow:     d.BillBoardNetAmt,
			BuySeats:      []*DragonTigerSeat{},
			SellSeats:     []*DragonTigerSeat{},
		})
	}
	return items, nil
}

type DragonTigerSeatResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Data []struct {
			OperatedeptName string  `json:"OPERATEDEPT_NAME"`
			NetAmt          float64 `json:"NET"`
			BuyAmt          float64 `json:"BUY"`
			SellAmt         float64 `json:"SELL"`
		} `json:"data"`
	} `json:"result"`
}

func (c *Client) GetDragonTigerSeats(ctx context.Context, code, date string) ([]*DragonTigerSeat, []*DragonTigerSeat, error) {
	urlBuy := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_BILLBOARD_DAILYDETAILSBUY&columns=ALL&filter=(SECURITY_CODE=%%22%s%%22)(TRADE_DATE=%%27%s%%27)", code, date)
	buySeats, err := c.fetchSeats(ctx, urlBuy)
	if err != nil {
		return nil, nil, err
	}

	urlSell := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_BILLBOARD_DAILYDETAILSSELL&columns=ALL&filter=(SECURITY_CODE=%%22%s%%22)(TRADE_DATE=%%27%s%%27)", code, date)
	sellSeats, err := c.fetchSeats(ctx, urlSell)
	if err != nil {
		return nil, nil, err
	}

	return buySeats, sellSeats, nil
}

func (c *Client) fetchSeats(ctx context.Context, url string) ([]*DragonTigerSeat, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result DragonTigerSeatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	var seats []*DragonTigerSeat
	for _, s := range result.Result.Data {
		seats = append(seats, &DragonTigerSeat{
			Name:    s.OperatedeptName,
			BuyAmt:  s.BuyAmt,
			SellAmt: s.SellAmt,
			NetAmt:  s.NetAmt,
			Tags:    []string{},
		})
	}
	return seats, nil
}

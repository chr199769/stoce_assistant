package eastmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"stock_assistant/backend/stock_service/kitex_gen/stock"
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

func (c *Client) GetFinancialReports(ctx context.Context, code string) ([]*stock.FinancialData, error) {
	// Clean code (remove sh/sz prefix if present, as EastMoney uses 6 digit code in filter)
	cleanCode := code
	if len(code) > 6 {
		cleanCode = code[len(code)-6:]
	}

	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_LICO_FN_CPD&columns=ALL&filter=(SECURITY_CODE=%%22%s%%22)&pageNumber=1&pageSize=5&sortTypes=-1&sortColumns=REPORTDATE", cleanCode)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
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
		// It might be successful but with empty data if code is wrong or no data
		if len(result.Result.Data) == 0 {
			return []*stock.FinancialData{}, nil
		}
		return nil, fmt.Errorf("eastmoney api failed")
	}

	var reports []*stock.FinancialData
	for _, item := range result.Result.Data {
		// ReportDate is usually "2024-03-31 00:00:00", we might want just the date part
		date := item.ReportDate
		if len(date) > 10 {
			date = date[:10]
		}

		reports = append(reports, &stock.FinancialData{
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

// --- Sector Data Support ---

type SectorRankResponse struct {
	Rc   int    `json:"rc"`
	Data *struct {
		Total int `json:"total"`
		Diff  []struct {
			Code             string  `json:"f12"`
			Name             string  `json:"f14"`
			ChangePercent    float64 `json:"f3"`
			NetInflow        float64 `json:"f62"`
			TopStockName     string  `json:"f128"`
			TopStockCode     string  `json:"f140"`
			TopStockChange   float64 `json:"f136"`
		} `json:"diff"`
	} `json:"data"`
}

// GetSectorRank fetches the top sectors by change percent.
// rankType: "concept" or "industry"
func (c *Client) GetSectorRank(ctx context.Context, rankType string, limit int) ([]*SectorInfo, error) {
	if limit <= 0 {
		limit = 20
	}

	// Determine 'fs' parameter based on type
	var fs string
	if rankType == "industry" {
		fs = "m:90+t:2+f:!50"
	} else {
		// Default to concept
		fs = "m:90+t:3+f:!50"
	}

	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=%d&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=%s&fields=f12,f13,f14,f2,f3,f62,f128,f140,f136", limit, fs)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
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
		return nil, fmt.Errorf("no data returned from eastmoney")
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

type SectorInfo struct {
	Code          string
	Name          string
	ChangePercent float64
	NetInflow     float64
	TopStockName  string
	TopStockCode  string
}

// --- Sector Details Support ---

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

type SectorStockItem struct {
	Code          string
	Name          string
	Price         float64
	ChangePercent float64
	Volume        int64
	Amount        float64
	MarketCap     float64
}

func (c *Client) GetSectorStocksRaw(ctx context.Context, sectorCode string) ([]*SectorStockItem, error) {
	fs := fmt.Sprintf("b:%s", sectorCode)
	// Get top 100 stocks by change percent desc
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=100&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=%s&fields=f12,f14,f2,f3,f5,f6,f20", fs)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
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
		return nil, fmt.Errorf("no data returned from eastmoney")
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

// --- Dragon Tiger Support ---

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

// DragonTigerItem is an internal struct to match EastMoney Data
type DragonTigerItem struct {
	Code          string
	Name          string
	ClosePrice    float64
	ChangePercent float64
	Reason        string
	NetInflow     float64
	BuySeats      []*DragonTigerSeat
	SellSeats     []*DragonTigerSeat
}

type DragonTigerSeat struct {
	Name    string
	BuyAmt  float64
	SellAmt float64
	NetAmt  float64
	Tags    []string
}

func (c *Client) GetDragonTigerList(ctx context.Context, date string) ([]*DragonTigerItem, error) {
	// RPT_DAILYBILLBOARD_DETAILS
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_DAILYBILLBOARD_DETAILS&columns=ALL&filter=(TRADE_DATE=%%27%s%%27)&pageNumber=1&pageSize=100&sortTypes=-1&sortColumns=BILLBOARD_NET_AMT", date)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
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
		return nil, fmt.Errorf("eastmoney api failed or no data")
	}

	var items []*DragonTigerItem
	for _, d := range result.Result.Data {
		items = append(items, &DragonTigerItem{
			Code:          d.SecurityCode,
			Name:          d.SecurityName,
			ClosePrice:    d.ClosePrice,
			ChangePercent: d.ChangeRate,
			Reason:        d.Explain,
			NetInflow:     d.BillBoardNetAmt,
			BuySeats:      []*DragonTigerSeat{}, // Populated later
			SellSeats:     []*DragonTigerSeat{},
		})
	}
	return items, nil
}

func (c *Client) GetDragonTigerSeats(ctx context.Context, code, date string) ([]*DragonTigerSeat, []*DragonTigerSeat, error) {
	// Buy Seats
	urlBuy := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_BILLBOARD_DAILYDETAILSBUY&columns=ALL&filter=(SECURITY_CODE=%%22%s%%22)(TRADE_DATE=%%27%s%%27)", code, date)
	buySeats, err := c.fetchSeats(ctx, urlBuy)
	if err != nil {
		return nil, nil, err
	}

	// Sell Seats
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
	resp, err := c.httpClient.Do(req)
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
			Tags:    []string{}, // Populated by logic
		})
	}
	return seats, nil
}

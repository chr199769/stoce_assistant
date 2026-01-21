package eastmoney

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Internal response structs

type dragonTigerStatusResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Data []struct {
			SecurityCode string  `json:"SECURITY_CODE"`
			SecurityName string  `json:"SECURITY_NAME_ABBR"`
			Explain      string  `json:"EXPLANATION"`
			ClosePrice   float64 `json:"CLOSE_PRICE"`
			ChangeRate   float64 `json:"CHANGE_RATE"`
		} `json:"data"`
	} `json:"result"`
}

func (c *Client) GetDragonTigerStatus(ctx context.Context, code string) (bool, string, error) {
	cleanCode := getCleanCode(code)
	url := "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_LHB_YZLIST&columns=ALL&pageNumber=1&pageSize=500&sortTypes=-1&sortColumns=Turnover"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, "", err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", err
	}

	var result dragonTigerStatusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, "", err
	}

	if !result.Success || len(result.Result.Data) == 0 {
		return false, "", nil
	}

	for _, item := range result.Result.Data {
		if item.SecurityCode == cleanCode {
			return true, fmt.Sprintf("On List: %s (Change: %.2f%%)", item.Explain, item.ChangeRate), nil
		}
	}

	return false, "", nil
}

type stockNewsResponse struct {
	Result struct {
		Data []struct {
			Title    string `json:"title"`
			ShowTime string `json:"show_time"`
			Url      string `json:"url"`
			Summary  string `json:"summary"`
		} `json:"data"`
	} `json:"result"`
}

func (c *Client) GetStockNews(ctx context.Context, code string) ([]*StockNewsItem, error) {
	cleanCode := getCleanCode(code)
	url := fmt.Sprintf("https://search-api-web.eastmoney.com/search/jsonp/news/list?param=(code=%s&p=1&ps=5)", cleanCode)

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

	// Handle JSONP if necessary? The URL says jsonp but usually returns JSON if no callback provided?
	// The original code handled JSONP.
	content := string(body)
	if idx := strings.Index(content, "("); idx != -1 && strings.HasSuffix(strings.TrimSpace(content), ")") {
		content = content[idx+1 : strings.LastIndex(content, ")")]
	}

	var result stockNewsResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}

	var news []*StockNewsItem
	for _, item := range result.Result.Data {
		title := strings.ReplaceAll(item.Title, "<em>", "")
		title = strings.ReplaceAll(title, "</em>", "")
		news = append(news, &StockNewsItem{
			Title:    title,
			ShowTime: item.ShowTime,
			Url:      item.Url,
			Summary:  item.Summary,
		})
	}
	return news, nil
}

type stockHeatResponse struct {
	Data []struct {
		SecurityCode string `json:"securityCode"`
		SecurityName string `json:"securityName"`
		Rank         int    `json:"rank"`
		Heat         int    `json:"heat"`
	} `json:"data"`
}

func (c *Client) GetStockHeat(ctx context.Context, code string) (*StockHeatData, error) {
	cleanCode := getCleanCode(code)
	url := "https://gbcdn.dfcfw.com/rank/popularityList.js"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes)

	prefix := "var popularityList='"
	if idx := strings.Index(body, prefix); idx != -1 {
		body = body[idx+len(prefix):]
	}
	if idx := strings.LastIndex(body, "'"); idx != -1 {
		body = body[:idx]
	}

	decoded, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %v", err)
	}

	var jsonData []byte
	// Try Zlib
	r, err := zlib.NewReader(bytes.NewReader(decoded))
	if err == nil {
		jsonData, _ = io.ReadAll(r)
		r.Close()
	}

	// Try Flate
	if len(jsonData) == 0 {
		fr := flate.NewReader(bytes.NewReader(decoded))
		jsonData, _ = io.ReadAll(fr)
		fr.Close()
	}

	// Try Gzip
	if len(jsonData) == 0 {
		gr, err := gzip.NewReader(bytes.NewReader(decoded))
		if err == nil {
			jsonData, _ = io.ReadAll(gr)
			gr.Close()
		}
	}

	if len(jsonData) == 0 {
		return nil, fmt.Errorf("decompression failed")
	}

	// Parse JSON
	var heatResp stockHeatResponse
	if err := json.Unmarshal(jsonData, &heatResp); err == nil && len(heatResp.Data) > 0 {
		for _, item := range heatResp.Data {
			if item.SecurityCode == cleanCode {
				return &StockHeatData{Rank: item.Rank, Heat: item.Heat}, nil
			}
		}
		return nil, nil // Not in top list
	}

	// Try array format
	var heatList []struct {
		SecurityCode string `json:"securityCode"`
		SecurityName string `json:"securityName"`
		Rank         int    `json:"rank"`
		Heat         int    `json:"heat"`
	}
	if err := json.Unmarshal(jsonData, &heatList); err == nil && len(heatList) > 0 {
		for _, item := range heatList {
			if item.SecurityCode == cleanCode {
				return &StockHeatData{Rank: item.Rank, Heat: item.Heat}, nil
			}
		}
		return nil, nil
	}

	return nil, fmt.Errorf("failed to parse heat data")
}

type noticeResponse struct {
	Data struct {
		List []struct {
			Title      string `json:"title"`
			Date       string `json:"notice_date"`
			ContentUrl string `json:"art_code"`
			Columns    []struct {
				Name string `json:"column_name"`
			} `json:"columns"`
		} `json:"list"`
	} `json:"data"`
}

func (c *Client) GetStockNotices(ctx context.Context, code string, filterKeywords []string) ([]*NoticeItem, error) {
	cleanCode := getCleanCode(code)
	url := fmt.Sprintf("https://np-anotice-stock.eastmoney.com/api/security/ann?sr=-1&page_size=20&page_index=1&ann_type=A&client_source=web&stock_list=%s", cleanCode)

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

	var result noticeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var notices []*NoticeItem
	for _, item := range result.Data.List {
		matched := len(filterKeywords) == 0
		if !matched {
			for _, kw := range filterKeywords {
				if strings.Contains(item.Title, kw) {
					matched = true
					break
				}
			}
			if !matched {
				for _, col := range item.Columns {
					for _, kw := range filterKeywords {
						if strings.Contains(col.Name, kw) {
							matched = true
							break
						}
					}
					if matched {
						break
					}
				}
			}
		}

		if matched {
			link := fmt.Sprintf("http://data.eastmoney.com/notices/detail/%s/%s.html", cleanCode, item.ContentUrl)
			date := item.Date
			if len(date) > 10 {
				date = date[:10]
			}
			notices = append(notices, &NoticeItem{
				Title: item.Title,
				Date:  date,
				Url:   link,
			})
		}
	}
	return notices, nil
}

type dragonTigerHistoryResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Data []struct {
			SecurityCode     string  `json:"SECURITY_CODE"`
			SecurityName     string  `json:"SECURITY_NAME_ABBR"`
			TradeDate        string  `json:"TRADE_DATE"`
			Explain          string  `json:"EXPLANATION"`
			ClosePrice       float64 `json:"CLOSE_PRICE"`
			ChangeRate       float64 `json:"CHANGE_RATE"`
			BillBoardNetAmt  float64 `json:"BILLBOARD_NET_AMT"`
			BillBoardBuyAmt  float64 `json:"BILLBOARD_BUY_AMT"`
			BillBoardSellAmt float64 `json:"BILLBOARD_SELL_AMT"`
			TurnoverRate     float64 `json:"TURNOVERRATE"`
		} `json:"data"`
	} `json:"result"`
}

func (c *Client) GetDragonTigerHistory(ctx context.Context, code string, limit int) ([]*DragonTigerItem, error) {
	cleanCode := getCleanCode(code)
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_DAILYBILLBOARD_DETAILS&columns=ALL&filter=(SECURITY_CODE=%%22%s%%22)&pageNumber=1&pageSize=%d&sortTypes=-1&sortColumns=TRADE_DATE", cleanCode, limit)

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

	var result dragonTigerHistoryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return []*DragonTigerItem{}, nil
	}

	var items []*DragonTigerItem
	for _, d := range result.Result.Data {
		date := d.TradeDate
		if len(date) > 10 {
			date = date[:10]
		}
		items = append(items, &DragonTigerItem{
			Date:          date,
			Code:          d.SecurityCode,
			Name:          d.SecurityName,
			ClosePrice:    d.ClosePrice,
			ChangePercent: d.ChangeRate,
			Reason:        d.Explain,
			NetInflow:     d.BillBoardNetAmt,
		})
	}
	return items, nil
}

type orderBookResponse struct {
	Data struct {
		Buy1Price float64 `json:"f19"`
		Buy1Vol   int     `json:"f20"`
		Buy2Price float64 `json:"f17"`
		Buy2Vol   int     `json:"f18"`
		Buy3Price float64 `json:"f15"`
		Buy3Vol   int     `json:"f16"`
		Buy4Price float64 `json:"f13"`
		Buy4Vol   int     `json:"f14"`
		Buy5Price float64 `json:"f11"`
		Buy5Vol   int     `json:"f12"`

		Sell1Price float64 `json:"f39"`
		Sell1Vol   int     `json:"f40"`
		Sell2Price float64 `json:"f37"`
		Sell2Vol   int     `json:"f38"`
		Sell3Price float64 `json:"f35"`
		Sell3Vol   int     `json:"f36"`
		Sell4Price float64 `json:"f33"`
		Sell4Vol   int     `json:"f34"`
		Sell5Price float64 `json:"f31"`
		Sell5Vol   int     `json:"f32"`

		WeiBi  interface{} `json:"f191"`
		WeiCha interface{} `json:"f192"`
	} `json:"data"`
}

func (c *Client) GetOrderBook(ctx context.Context, code string) (*OrderBookData, error) {
	secId := getSecId(code)
	url := fmt.Sprintf("http://push2.eastmoney.com/api/qt/stock/get?fltt=2&invt=2&klt=101&secid=%s&fields=f19,f20,f17,f18,f15,f16,f13,f14,f11,f12,f39,f40,f37,f38,f35,f36,f33,f34,f31,f32,f191,f192", secId)

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

	var result orderBookResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	data := &OrderBookData{
		Buy1Price:  result.Data.Buy1Price,
		Buy1Vol:    result.Data.Buy1Vol,
		Buy2Price:  result.Data.Buy2Price,
		Buy2Vol:    result.Data.Buy2Vol,
		Buy3Price:  result.Data.Buy3Price,
		Buy3Vol:    result.Data.Buy3Vol,
		Buy4Price:  result.Data.Buy4Price,
		Buy4Vol:    result.Data.Buy4Vol,
		Buy5Price:  result.Data.Buy5Price,
		Buy5Vol:    result.Data.Buy5Vol,
		Sell1Price: result.Data.Sell1Price,
		Sell1Vol:   result.Data.Sell1Vol,
		Sell2Price: result.Data.Sell2Price,
		Sell2Vol:   result.Data.Sell2Vol,
		Sell3Price: result.Data.Sell3Price,
		Sell3Vol:   result.Data.Sell3Vol,
		Sell4Price: result.Data.Sell4Price,
		Sell4Vol:   result.Data.Sell4Vol,
		Sell5Price: result.Data.Sell5Price,
		Sell5Vol:   result.Data.Sell5Vol,
	}

	// Handle WeiBi/WeiCha interface{}
	if val, ok := result.Data.WeiBi.(float64); ok {
		data.WeiBi = val
	}
	if val, ok := result.Data.WeiCha.(float64); ok {
		data.WeiCha = val
	}

	return data, nil
}

type chipDistributionResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Data []struct {
			SecurityCode string  `json:"SECURITY_CODE"`
			WinnerRate   float64 `json:"WINNER_RATE"`
			Cost90Low    float64 `json:"COST_90_LOW"`
			Cost90High   float64 `json:"COST_90_HIGH"`
			AverageCost  float64 `json:"AVERAGE_COST"`
		} `json:"data"`
	} `json:"result"`
}

func (c *Client) GetChipDistribution(ctx context.Context, code string) (*ChipDistributionData, error) {
	cleanCode := getCleanCode(code)
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_CYQ_OUTPUT&columns=ALL&filter=(SECURITY_CODE=%%22%s%%22)", cleanCode)

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

	var result chipDistributionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.Success || len(result.Result.Data) == 0 {
		return nil, nil
	}

	d := result.Result.Data[0]
	return &ChipDistributionData{
		AverageCost: d.AverageCost,
		WinnerRate:  d.WinnerRate,
		Cost90Low:   d.Cost90Low,
		Cost90High:  d.Cost90High,
	}, nil
}

type industryResponse struct {
	Data struct {
		IndustryName string `json:"f127"`
		RegionName   string `json:"f128"`
		ConceptNames string `json:"f129"`
	} `json:"data"`
}

func (c *Client) GetIndustryIndex(ctx context.Context, code string) (*IndustryIndexData, error) {
	secId := getSecId(code)
	url := fmt.Sprintf("http://push2.eastmoney.com/api/qt/stock/get?secid=%s&fields=f127,f128,f129", secId)

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

	var result industryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &IndustryIndexData{
		IndustryName: result.Data.IndustryName,
		RegionName:   result.Data.RegionName,
		ConceptNames: result.Data.ConceptNames,
	}, nil
}

package tool

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Common HTTP client for EastMoney APIs
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// fetchJSON fetches JSON data from a URL and decodes it into the target struct
func fetchJSON(url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Handle JSONP if necessary (EastMoney often uses it)
	// Simple check: if starts with 'callback(' or similar
	content := string(body)
	if idx := strings.Index(content, "("); idx != -1 && strings.HasSuffix(strings.TrimSpace(content), ")") {
		content = content[idx+1 : strings.LastIndex(content, ")")]
	}

	return json.Unmarshal([]byte(content), target)
}

// --- Data Structures ---

// DragonTigerResponse matches https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_LHB_YZLIST...
type DragonTigerResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Data []struct {
			SecurityCode string  `json:"SECURITY_CODE"`
			SecurityName string  `json:"SECURITY_NAME_ABBR"`
			Explain      string  `json:"EXPLANATION"` // Reason for being on list
			ClosePrice   float64 `json:"CLOSE_PRICE"`
			ChangeRate   float64 `json:"CHANGE_RATE"`
		} `json:"data"`
	} `json:"result"`
}

// StockNewsResponse matches https://search-api-web.eastmoney.com/search/jsonp/news/list
type StockNewsResponse struct {
	Result struct {
		Data []struct {
			Title    string `json:"title"`
			ShowTime string `json:"show_time"`
			Url      string `json:"url"`
			Summary  string `json:"summary"`
		} `json:"data"`
	} `json:"result"`
}

// StockHeatResponse matches https://gbcdn.dfcfw.com/rank/popularityList.js
type StockHeatResponse struct {
	Data []struct {
		SecurityCode string `json:"securityCode"`
		SecurityName string `json:"securityName"`
		Rank         int    `json:"rank"`
		Heat         int    `json:"heat"`
	} `json:"data"`
}

// DragonTigerHistoryResponse matches RPT_LHB_INDIVIDUAL interface
type DragonTigerHistoryResponse struct {
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

// OrderBookResponse matches http://push2.eastmoney.com/api/qt/stock/get
type OrderBookResponse struct {
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

		WeiBi  interface{} `json:"f191"` // Wei Bi (Commission Ratio)
		WeiCha interface{} `json:"f192"` // Wei Cha (Commission Diff)
	} `json:"data"`
}

// ChipDistributionResponse matches https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_CYQ_OUTPUT...
type ChipDistributionResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Data []struct {
			SecurityCode string  `json:"SECURITY_CODE"`
			WinnerRate   float64 `json:"WINNER_RATE"`  // Profit ratio
			Cost90Low    float64 `json:"COST_90_LOW"`  // 90% cost range low
			Cost90High   float64 `json:"COST_90_HIGH"` // 90% cost range high
			AverageCost  float64 `json:"AVERAGE_COST"` // Average cost
		} `json:"data"`
	} `json:"result"`
}

// IndustryResponse matches http://push2.eastmoney.com/api/qt/stock/get?fields=f127,f128,f129
type IndustryResponse struct {
	Data struct {
		IndustryName string `json:"f127"` // Industry Name
		RegionName   string `json:"f128"` // Region Name
		ConceptNames string `json:"f129"` // Concept Names
	} `json:"data"`
}

// NorthboundResponse matches http://push2.eastmoney.com/api/qt/kamt/get
type NorthboundResponse struct {
	Data struct {
		HK2SH struct {
			DayNetAmtIn float64 `json:"f52"`
		} `json:"hk2sh"`
		HK2SZ struct {
			DayNetAmtIn float64 `json:"f52"`
		} `json:"hk2sz"`
	} `json:"data"`
}

// NoticeResponse matches https://np-anotice-stock.eastmoney.com/api/security/ann
type NoticeResponse struct {
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

// --- Fetch Functions ---

// GetDragonTigerStatus checks if a stock is on the latest Dragon & Tiger list
func GetDragonTigerStatus(code string) (bool, string, error) {
	// Clean code (remove sh/sz prefix for EastMoney query if needed, but response has SECURITY_CODE as 6 digits)
	cleanCode := code
	if len(code) > 6 {
		cleanCode = code[len(code)-6:]
	}

	// Fetch latest list (page 1, size 500 should cover most)
	url := "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_LHB_YZLIST&columns=ALL&pageNumber=1&pageSize=500&sortTypes=-1&sortColumns=Turnover"
	var resp DragonTigerResponse
	if err := fetchJSON(url, &resp); err != nil {
		return false, "", err
	}

	if !resp.Success || len(resp.Result.Data) == 0 {
		return false, "", nil
	}

	for _, item := range resp.Result.Data {
		if item.SecurityCode == cleanCode {
			return true, fmt.Sprintf("On List: %s (Change: %.2f%%)", item.Explain, item.ChangeRate), nil
		}
	}

	return false, "", nil
}

// GetStockNews fetches recent news for a stock
func GetStockNews(code string) ([]string, error) {
	cleanCode := code
	if len(code) > 6 {
		cleanCode = code[len(code)-6:]
	}

	url := fmt.Sprintf("https://search-api-web.eastmoney.com/search/jsonp/news/list?param=(code=%s&p=1&ps=5)", cleanCode)
	var resp StockNewsResponse
	if err := fetchJSON(url, &resp); err != nil {
		return nil, err
	}

	var news []string
	for _, item := range resp.Result.Data {
		// Remove HTML tags from title/summary if present
		title := strings.ReplaceAll(item.Title, "<em>", "")
		title = strings.ReplaceAll(title, "</em>", "")
		news = append(news, fmt.Sprintf("[%s] %s", item.ShowTime, title))
	}
	return news, nil
}

// --- New Fetch Functions for Alpha Signals ---

// GetNorthboundFunds fetches the real-time Northbound funds inflow
func GetNorthboundFunds() (string, error) {
	// f52 is Net Inflow
	url := "http://push2.eastmoney.com/api/qt/kamt/get?fields1=f1,f2,f3,f4&fields2=f51,f52,f53,f54"

	var resp NorthboundResponse
	if err := fetchJSON(url, &resp); err != nil {
		return "", err
	}

	shNet := resp.Data.HK2SH.DayNetAmtIn / 10000 // Convert to Wan
	szNet := resp.Data.HK2SZ.DayNetAmtIn / 10000 // Convert to Wan
	total := shNet + szNet

	return fmt.Sprintf("Northbound Net Inflow: Total %.0fWan (SH: %.0fWan, SZ: %.0fWan)", total, shNet, szNet), nil
}

// GetStockHeat fetches the popularity rank of a stock from Eastmoney Guba
func GetStockHeat(code string) (string, error) {
	// Stub implementation as the previous API is protected/encrypted
	return "Stock Heat Data: Unavailable (API Protected)", nil
}

// GetStockNotices fetches official announcements, including inquiries and regulation letters
// filterTypes: keywords to filter title or column, e.g., []string{"监管", "问询", "关注函"}
func GetStockNotices(code string, filterKeywords []string) ([]string, error) {
	cleanCode := code
	if len(code) > 6 {
		cleanCode = code[len(code)-6:]
	}

	// API returns JSON directly (not JSONP)
	// stock_list format: code (e.g., 600519)
	url := fmt.Sprintf("https://np-anotice-stock.eastmoney.com/api/security/ann?sr=-1&page_size=20&page_index=1&ann_type=A&client_source=web&stock_list=%s", cleanCode)

	var resp NoticeResponse
	if err := fetchJSON(url, &resp); err != nil {
		return nil, err
	}

	var notices []string
	for _, item := range resp.Data.List {
		// Check if item matches any keyword (if filters provided)
		matched := len(filterKeywords) == 0
		if !matched {
			// Check title
			for _, kw := range filterKeywords {
				if strings.Contains(item.Title, kw) {
					matched = true
					break
				}
			}
			// Check columns (tags)
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
			// Construct PDF/Page URL: http://data.eastmoney.com/notices/detail/{code}/{art_code}.html
			link := fmt.Sprintf("http://data.eastmoney.com/notices/detail/%s/%s.html", cleanCode, item.ContentUrl)
			notices = append(notices, fmt.Sprintf("[%s] %s", item.Date[:10], item.Title))
			// Only keep link if needed, but for prompt brevity, title is more important.
			// Add link if user wants to read details (maybe in future)
			_ = link
		}
	}
	return notices, nil
}

// --- New Fetch Functions for Optimization ---

// GetDragonTigerHistory fetches historical Dragon Tiger list records for a stock (last N records)
func GetDragonTigerHistory(code string, limit int) ([]string, error) {
	cleanCode := code
	if len(code) > 6 {
		cleanCode = code[len(code)-6:]
	}

	// Construct URL: sort by trade date descending, get latest limit records
	// Use %22 for quotes in filter
	// Changed RPT_LHB_INDIVIDUAL to RPT_LHB_HISTORYLIST
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_LHB_HISTORYLIST&columns=ALL&filter=(SECURITY_CODE=%%22%s%%22)&pageNumber=1&pageSize=%d&sortTypes=-1&sortColumns=TRADE_DATE", cleanCode, limit)

	var resp DragonTigerHistoryResponse
	if err := fetchJSON(url, &resp); err != nil {
		return nil, err
	}

	if !resp.Success || len(resp.Result.Data) == 0 {
		return nil, nil // No history
	}

	var history []string
	for _, item := range resp.Result.Data {
		// Format date, remove time part
		date := item.TradeDate
		if len(date) > 10 {
			date = date[:10]
		}

		netBuy := item.BillBoardNetAmt / 10000 // Convert to Wan

		record := fmt.Sprintf("[%s] %s | Change: %.2f%% | NetBuy: %.0fWan | Reason: %s",
			date, item.SecurityName, item.ChangeRate, netBuy, item.Explain)
		history = append(history, record)
	}

	return history, nil
}

// getSecId converts stock code to EastMoney secid format (1.6xxxxx for SH, 0.xxxxxx for SZ/BJ)
func getSecId(code string) string {
	cleanCode := code
	if len(code) > 6 {
		cleanCode = code[len(code)-6:]
	}
	// Simple rule: 6 starts for SH (1), others for SZ/BJ (0)
	if strings.HasPrefix(cleanCode, "6") {
		return "1." + cleanCode
	}
	return "0." + cleanCode
}

// GetOrderBook fetches the real-time 5-level order book
func GetOrderBook(code string) (string, error) {
	secId := getSecId(code)
	// f11-f40 covers buy/sell 5 levels prices and volumes
	url := fmt.Sprintf("http://push2.eastmoney.com/api/qt/stock/get?fltt=2&invt=2&klt=101&secid=%s&fields=f19,f20,f17,f18,f15,f16,f13,f14,f11,f12,f39,f40,f37,f38,f35,f36,f33,f34,f31,f32,f191,f192", secId)

	var resp OrderBookResponse
	if err := fetchJSON(url, &resp); err != nil {
		return "", err
	}

	// Format output
	var sb strings.Builder

	// Add WeiBi/WeiCha if available
	if resp.Data.WeiBi != nil && resp.Data.WeiCha != nil {
		sb.WriteString(fmt.Sprintf("Intraday Pressure: WeiBi(%.2f%%), WeiCha(%.0f) | ", resp.Data.WeiBi, resp.Data.WeiCha))
	}

	// Add Order Book Snapshot
	// Check if data is valid (sometimes returns 0s if market closed or data missing)
	if resp.Data.Buy1Price > 0 || resp.Data.Sell1Price > 0 {
		sb.WriteString(fmt.Sprintf("Buy1: %.2f(%d), Sell1: %.2f(%d)",
			resp.Data.Buy1Price, resp.Data.Buy1Vol,
			resp.Data.Sell1Price, resp.Data.Sell1Vol))
	} else {
		sb.WriteString("Order Book Snapshot: Unavailable (Market Closed or Level-1 Restricted)")
	}

	return sb.String(), nil
}

// GetChipDistribution fetches the chip distribution (cost concentration)
func GetChipDistribution(code string) (string, error) {
	cleanCode := code
	if len(code) > 6 {
		cleanCode = code[len(code)-6:]
	}

	// URL for CYQ (Cost Yield Curve) data
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_CYQ_OUTPUT&columns=ALL&filter=(SECURITY_CODE=%%22%s%%22)", cleanCode)

	var resp ChipDistributionResponse
	if err := fetchJSON(url, &resp); err != nil {
		return "", err
	}

	if !resp.Success || len(resp.Result.Data) == 0 {
		return "No chip distribution data available", nil
	}

	data := resp.Result.Data[0]
	return fmt.Sprintf("AvgCost: %.2f, WinnerRate: %.2f%%, 90%%CostRange: %.2f-%.2f",
		data.AverageCost, data.WinnerRate, data.Cost90Low, data.Cost90High), nil
}

// GetIndustryIndex fetches the industry sector info
func GetIndustryIndex(code string) (string, error) {
	secId := getSecId(code)
	url := fmt.Sprintf("http://push2.eastmoney.com/api/qt/stock/get?secid=%s&fields=f127,f128,f129", secId)

	var resp IndustryResponse
	if err := fetchJSON(url, &resp); err != nil {
		return "", err
	}

	return fmt.Sprintf("Industry: %s, Region: %s, Concepts: %s",
		resp.Data.IndustryName, resp.Data.RegionName, resp.Data.ConceptNames), nil
}

// GetMarketNews fetches general market news (kuaixun)
func GetMarketNews() ([]string, error) {
	// This URL returns a list of news objects
	url := "https://newsapi.eastmoney.com/kuaixun/v1/getlist_102_0_0_20.html"

	// The structure is slightly different for Kuaixun
	type KuaixunResponse struct {
		Lives []struct {
			Title  string `json:"title"`
			Digest string `json:"digest"`
			Time   string `json:"showtime"`
		} `json:"Lives"`
	}

	var resp KuaixunResponse
	if err := fetchJSON(url, &resp); err != nil {
		return nil, err
	}

	var news []string
	for _, item := range resp.Lives {
		content := item.Digest
		if content == "" {
			content = item.Title
		}
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		news = append(news, fmt.Sprintf("[%s] %s", item.Time, content))
	}
	return news, nil
}

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

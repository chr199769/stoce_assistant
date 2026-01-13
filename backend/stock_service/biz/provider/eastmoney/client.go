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

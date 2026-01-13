package sina

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stock_assistant/backend/stock_service/kitex_gen/stock"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Client handles interaction with Sina Finance API
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Sina Finance API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetStockInfo fetches real-time stock information
// code format: sh600000, sz000001
func (c *Client) GetStockInfo(ctx context.Context, code string) (*stock.StockInfo, error) {
	// Auto-fix code prefix if missing
	code = strings.TrimSpace(code)
	if len(code) == 6 {
		if strings.HasPrefix(code, "6") {
			code = "sh" + code
		} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
			code = "sz" + code
		}
	}

	// Sina API format: http://hq.sinajs.cn/list=sh601006
	url := fmt.Sprintf("http://hq.sinajs.cn/list=%s", code)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	// Set headers to mimic a browser to avoid potential blocking
	req.Header.Set("Referer", "https://finance.sina.com.cn/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	// Sina API returns GBK/GB18030 encoding, need to convert to UTF-8
	// Read raw bytes first to avoid reader error on invalid chars
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var content string
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8Body, _, err := transform.Bytes(decoder, rawBody)
	if err != nil {
		// If decoding fails (e.g. GB18030 specific chars), try to use raw body if it's mostly ASCII compatible for numbers
		// Or try to ignore errors?
		// For now, let's log and fallback to treating it as string (might be garbled for name but numbers work)
		fmt.Printf("GBK decoding failed for %s: %v. Using raw string.\n", code, err)
		content = string(rawBody)
	} else {
		content = string(utf8Body)
	}
	// Response format: var hq_str_sh601006="大秦铁路,6.660,6.660,6.670,6.690,6.640,6.660,6.670,22759685,151717808.000,189400,6.660,119900,6.650,229400,6.640,147800,6.630,121700,6.620,100,6.670,165700,6.680,187800,6.690,266600,6.700,240200,6.710,2024-05-15,15:00:00,00,";

	// Check if data is valid
	if !strings.Contains(content, "=\"") {
		return nil, fmt.Errorf("invalid stock code or empty response. Content: %s, URL: %s", content, url)
	}

	parts := strings.Split(content, "=\"")
	if len(parts) < 2 {
		return nil, fmt.Errorf("parse error. Content: %s", content)
	}

	dataStr := strings.TrimSuffix(parts[1], "\";\n")
	// Some response might end with "; without newline or just ";
	dataStr = strings.TrimSuffix(dataStr, "\";")

	if dataStr == "" {
		return nil, fmt.Errorf("empty data")
	}

	fields := strings.Split(dataStr, ",")
	// Index data (sh000001) has fewer fields than stock data, usually around 6-10 fields depending on status
	// But Sina usually returns 32+ fields for individual stocks.
	// For indices like sh000001, it returns:
	// "上证指数,3026.9877,3038.2259,3021.3533,3044.6970,3016.5367,0,0,323049586,346985022378,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,2024-05-15,15:30:00,00,";
	// The length check should be more lenient or specific to the type.

	// A basic valid response should have at least name, prices and date/time.
	if len(fields) < 6 {
		return nil, fmt.Errorf("unexpected data format: fields count %d. Data: %s, URL: %s, Content: %s", len(fields), dataStr, url, content)
	}

	// Parse fields
	name := fields[0]
	openPrice, _ := strconv.ParseFloat(fields[1], 64)
	_ = openPrice // prevClose is fields[2]
	prevClose, _ := strconv.ParseFloat(fields[2], 64)
	currentPrice, _ := strconv.ParseFloat(fields[3], 64)
	// fields[4] is high, fields[5] is low
	// fields[8] is volume (shares)
	volume, _ := strconv.ParseInt(fields[8], 10, 64)
	date := fields[30]
	timeStr := fields[31]

	// Calculate change percent
	changePercent := 0.0
	if prevClose > 0 {
		changePercent = (currentPrice - prevClose) / prevClose * 100
	}

	return &stock.StockInfo{
		Code:          code,
		Name:          name,
		CurrentPrice:  currentPrice,
		ChangePercent: changePercent,
		Volume:        volume,
		Timestamp:     fmt.Sprintf("%s %s", date, timeStr),
	}, nil
}

package eastmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// --- Kline History Support ---

type KlineResponse struct {
	Data *struct {
		Code   string   `json:"code"`
		Name   string   `json:"name"`
		Klines []string `json:"klines"` // Format: "2024-01-01,10.0,11.0,9.0,10.5,100000,..."
	} `json:"data"`
}

type KlineItem struct {
	Date   string
	Open   float64
	Close  float64
	High   float64
	Low    float64
	Volume int64
}

func (c *Client) GetKlineHistory(ctx context.Context, code string, days int) ([]*KlineItem, error) {
	// 1. Determine market code (0 for SZ, 1 for SH)
	// Input code might be "sh600519" or "600519"
	// Simple heuristic: 6xxxx -> 1, others -> 0 (Not 100% accurate but works for main boards)
	// Better: check prefix
	secID := "0." + code
	if len(code) > 2 {
		prefix := code[:2]
		numCode := code
		if prefix == "sh" || prefix == "sz" {
			numCode = code[2:]
		}
		
		if prefix == "sh" {
			secID = "1." + numCode
		} else if prefix == "sz" {
			secID = "0." + numCode
		} else {
			// Guess based on first digit
			if numCode[0] == '6' {
				secID = "1." + numCode
			} else {
				secID = "0." + numCode
			}
		}
	}

	// Limit calculation is tricky with "days", EastMoney uses beg/end date or limit count
	// lmt=days
	url := fmt.Sprintf("http://push2his.eastmoney.com/api/qt/stock/kline/get?fields1=f1&fields2=f51,f52,f53,f54,f55,f56&klt=101&fqt=1&secid=%s&lmt=%d&end=20500101", secID, days)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Use standard browser headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "http://quote.eastmoney.com/")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")

	// We use c.httpClient directly here to avoid the middleware in c.doRequest that might force HTTPS or add incompatible headers for this specific endpoint if any
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eastmoney kline api error: status=%d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result KlineResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Data == nil {
		return nil, fmt.Errorf("no kline data")
	}

	var klines []*KlineItem
	for _, kStr := range result.Data.Klines {
		// Format: "2024-03-20,10.1,10.2,10.0,10.15,50000,..."
		// Fields: Date, Open, Close, High, Low, Volume
		parts := strings.Split(kStr, ",")
		if len(parts) < 6 {
			continue
		}
		
		open, _ := strconv.ParseFloat(parts[1], 64)
		close, _ := strconv.ParseFloat(parts[2], 64)
		high, _ := strconv.ParseFloat(parts[3], 64)
		low, _ := strconv.ParseFloat(parts[4], 64)
		vol, _ := strconv.ParseInt(parts[5], 10, 64)

		klines = append(klines, &KlineItem{
			Date:   parts[0],
			Open:   open,
			Close:  close,
			High:   high,
			Low:    low,
			Volume: vol,
		})
	}

	return klines, nil
}

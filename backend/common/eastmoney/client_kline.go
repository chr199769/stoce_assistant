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

type KlineResponse struct {
	Data *struct {
		Code   string   `json:"code"`
		Name   string   `json:"name"`
		Klines []string `json:"klines"`
	} `json:"data"`
}

func (c *Client) GetKlineHistory(ctx context.Context, code string, days int) ([]*KlineItem, error) {
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
			if numCode[0] == '6' {
				secID = "1." + numCode
			} else {
				secID = "0." + numCode
			}
		}
	}

	url := fmt.Sprintf("http://push2his.eastmoney.com/api/qt/stock/kline/get?fields1=f1&fields2=f51,f52,f53,f54,f55,f56,f59&klt=101&fqt=1&secid=%s&lmt=%d&end=20500101", secID, days)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "http://quote.eastmoney.com/")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")

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
		parts := strings.Split(kStr, ",")
		if len(parts) < 7 {
			continue
		}
		
		open, _ := strconv.ParseFloat(parts[1], 64)
		close, _ := strconv.ParseFloat(parts[2], 64)
		high, _ := strconv.ParseFloat(parts[3], 64)
		low, _ := strconv.ParseFloat(parts[4], 64)
		vol, _ := strconv.ParseInt(parts[5], 10, 64)
		changePct, _ := strconv.ParseFloat(parts[6], 64)

		klines = append(klines, &KlineItem{
			Date:          parts[0],
			Open:          open,
			Close:         close,
			High:          high,
			Low:           low,
			Volume:        vol,
			ChangePercent: changePct,
		})
	}

	return klines, nil
}

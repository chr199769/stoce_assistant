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
	secID := getSecId(code)
	hosts := []string{
		"https://push2his.eastmoney.com",
		"https://push2.eastmoney.com",
	}
	// 说明：多主机轮询以提升可用性；服务端偶发 EOF/限流时快速切换
	var lastErr error
	for i, host := range hosts {
		url := fmt.Sprintf("%s/api/qt/stock/kline/get?fields1=f1&fields2=f51,f52,f53,f54,f55,f56,f59&klt=101&fqt=1&secid=%s&lmt=%d&end=20500101", host, secID, days)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		resp, err := c.doRequest(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}
		var result KlineResponse
		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = err
			continue
		}
		if result.Data == nil || len(result.Data.Klines) == 0 {
			lastErr = fmt.Errorf("no kline data")
			continue
		}
		return parseKlineItems(result.Data.Klines), nil
		_ = i
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("kline fetch failed")
}

func (c *Client) GetKlineHistoryWithKlt(ctx context.Context, code string, limit int, klt int) ([]*KlineItem, error) {
	secID := getSecId(code)
	hosts := []string{
		"https://push2his.eastmoney.com",
		"https://push2.eastmoney.com",
	}
	// 说明：多主机轮询以提升可用性；klt 周期由适配层映射
	var lastErr error
	for _, host := range hosts {
		url := fmt.Sprintf("%s/api/qt/stock/kline/get?fields1=f1&fields2=f51,f52,f53,f54,f55,f56,f59&klt=%d&fqt=1&secid=%s&lmt=%d&end=20500101", host, klt, secID, limit)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		resp, err := c.doRequest(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}
		var result KlineResponse
		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = err
			continue
		}
		if result.Data == nil || len(result.Data.Klines) == 0 {
			lastErr = fmt.Errorf("no kline data")
			continue
		}
		return parseKlineItems(result.Data.Klines), nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("kline fetch failed")
}

func parseKlineItems(raw []string) []*KlineItem {
	var klines []*KlineItem
	for _, kStr := range raw {
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
	return klines
}

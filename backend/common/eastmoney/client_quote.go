package eastmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type realtimeQuoteResponse struct {
	Data struct {
		Code      string  `json:"f57"`
		Name      string  `json:"f58"`
		// f43: 最新价
		Current   float64 `json:"f43"`
		// f44: 最高
		High      float64 `json:"f44"`
		// f45: 最低
		Low       float64 `json:"f45"`
		// f46: 昨收
		PrevClose float64 `json:"f46"`
		// f47: 开盘
		Open      float64 `json:"f47"`
		// f48: 成交额
		Amount    float64 `json:"f48"`
		// f49: 成交量
		Volume    float64 `json:"f49"`
		// f86: 时间（字符串）
		Time      string  `json:"f86"`
	} `json:"data"`
}

func (c *Client) GetRealtimeQuote(ctx context.Context, code string) (*RealtimeQuoteData, error) {
	secId := getSecId(code)
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/stock/get?secid=%s&fltt=2&invt=2&fields=f43,f44,f45,f46,f47,f48,f49,f57,f58,f86", secId)
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
	var result realtimeQuoteResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	d := result.Data
	return &RealtimeQuoteData{
		Code:      d.Code,
		Name:      d.Name,
		Current:   d.Current,
		PrevClose: d.PrevClose,
		Open:      d.Open,
		High:      d.High,
		Low:       d.Low,
		Volume:    int64(d.Volume),
		Amount:    d.Amount,
		Timestamp: d.Time,
	}, nil
}

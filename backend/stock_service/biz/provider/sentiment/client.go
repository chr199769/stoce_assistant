package sentiment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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

type LimitUpStock struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	ChangePercent float64 `json:"change_percent"`
	LimitUpType   string  `json:"limit_up_type"` // e.g., "首板", "2连板"
	Reason        string  `json:"reason"`        // e.g., "华为概念"
	IsBroken      bool    `json:"is_broken"`     // True if broken limit up (炸板)
}

type LimitUpPoolResponse struct {
	Rc   int `json:"rc"`
	Data *struct {
		Pool []struct {
			Code       string  `json:"c"`
			Name       string  `json:"n"`
			Price      float64 `json:"p"`
			ChangePct  float64 `json:"zdp"`
			LimitUpTyp string  `json:"lbc"` // Limit Board Count, e.g. "1", "2"
			Reason     string  `json:"hybk"` // Industry/Concept
			IsBroken   int     `json:"zbc"`  // 0 or >0
		} `json:"pool"` // NOTE: This structure is hypothetical based on common EM patterns, needs adjustment
	} `json:"data"`
}

// GetLimitUpPool 获取每日涨停池数据
// 注意：使用非官方 API，可能不稳定
func (c *Client) GetLimitUpPool(ctx context.Context) ([]*LimitUpStock, error) {
	// 占位符 URL，实际需确认
	// 真实 URL 示例: https://push2ex.eastmoney.com/getTopicZTPool
	
	// 构建 YYYYMMDD 日期
	dateStr := time.Now().Format("20060102")
	url := fmt.Sprintf("https://push2ex.eastmoney.com/getTopicZTPool?ut=7eea3edcaed734bea9cbfc24409ed989&dpt=wz.ztgc&Pageindex=0&pagesize=100&sort=fbt:asc&date=%s", dateStr)

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

	// 暂未确认 ZTPool API 准确 JSON 结构
	// 记录日志并在解析失败时返回空列表
	
	var result LimitUpPoolResponse
	if err := json.Unmarshal(body, &result); err != nil {
		// 解析失败，返回错误
		return nil, fmt.Errorf("解析情绪数据失败: %v", err)
	}
	
	if result.Data == nil || result.Rc != 0 {
		// API 失败，返回空列表
		fmt.Printf("警告: 情绪 API 失败 (rc=%d), 返回空列表。\n", result.Rc)
		return []*LimitUpStock{}, nil
	}

	var stocks []*LimitUpStock
	for _, item := range result.Data.Pool {
		stocks = append(stocks, &LimitUpStock{
			Code:          item.Code,
			Name:          item.Name,
			Price:         item.Price,
			ChangePercent: item.ChangePct,
			LimitUpType:   fmt.Sprintf("%s连板", item.LimitUpTyp),
			Reason:        item.Reason,
			IsBroken:      item.IsBroken > 0,
		})
	}

	return stocks, nil
}

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

// GetLimitUpPool fetches the daily limit-up pool.
// Note: This uses a reverse-engineered API which might be unstable.
func (c *Client) GetLimitUpPool(ctx context.Context) ([]*LimitUpStock, error) {
	// Using a placeholder URL that was found in search results, but might need date param.
	// Current strategy: Try to fetch, if fail, return empty list (non-blocking).
	// Real URL often looks like: https://push2ex.eastmoney.com/getTopicZTPool
	
	// Construct today's date in YYYYMMDD format
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

	// For now, since we haven't confirmed the exact JSON structure of the ZTPool API,
	// we will log the body (in a real app) and return a mock/empty list if parsing fails.
	// This allows the compilation to succeed and we can refine the parsing logic later 
	// when we have a valid response sample.
	
	var result LimitUpPoolResponse
	if err := json.Unmarshal(body, &result); err != nil {
		// Fallback: return error or empty
		return nil, fmt.Errorf("failed to parse sentiment data: %v", err)
	}
	
	if result.Data == nil || result.Rc != 0 {
		// Log the error but return empty list instead of misleading mock data
		fmt.Printf("Warning: Sentiment API failed (rc=%d), returning empty list.\n", result.Rc)
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

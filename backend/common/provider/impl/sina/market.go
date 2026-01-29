package sinaimpl

import (
	"context"
	"strings"
	"time"

	"stock_assistant/backend/common/provider"
	ss "stock_assistant/backend/stock_service/biz/provider/sina"
)

// Market 实现 MarketDataClient（仅 IntradayQuote）
type Market struct {
	client *ss.Client
}

func NewMarket(c *ss.Client) *Market {
	return &Market{client: c}
}

// GetIntradayQuote 使用新浪实时行情
func (m *Market) GetIntradayQuote(ctx context.Context, code string) (*provider.IntradayQuote, error) {
	info, err := m.client.GetStockInfo(ctx, code)
	if err != nil {
		return nil, err
	}
	// info.Timestamp 格式 "YYYY-MM-DD HH:MM:SS"
	ts := time.Now()
	if info.Timestamp != "" {
		layouts := []string{"2006-01-02 15:04:05", "2006-01-02 15:04"}
		for _, ly := range layouts {
			if t, e := time.Parse(ly, info.Timestamp); e == nil {
				ts = t
				break
			}
		}
	}
	name := info.Name
	// 兼容 name 可能含空格或特殊字符
	name = strings.TrimSpace(name)
	return &provider.IntradayQuote{
		Code:      info.Code,
		Name:      name,
		Price:     info.CurrentPrice,
		PrevClose: 0,
		Open:      0,
		High:      0,
		Low:       0,
		ChangePct: info.ChangePercent,
		Volume:    info.Volume,
		Amount:    0,
		Timestamp: ts,
	}, nil
}

// GetOrderBook 未实现
func (m *Market) GetOrderBook(ctx context.Context, code string, depth int) (*provider.OrderBook, error) {
	return nil, providerErr("sina", "GetOrderBook 未实现")
}

// GetKlines 未实现
func (m *Market) GetKlines(ctx context.Context, code string, period string, limit int) ([]*provider.KlineItem, error) {
	return nil, providerErr("sina", "GetKlines 未实现")
}

func providerErr(src, msg string) error {
	return &ProviderError{Source: src, Message: msg}
}

type ProviderError struct {
	Source  string
	Message string
}

func (e *ProviderError) Error() string {
	return e.Source + ": " + e.Message
}

package eastmoneyimpl

import (
	"context"
	"fmt"
	"time"

	em "stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/common/netease"
	"stock_assistant/backend/common/provider"
)

// EMClient 适配所需的东财客户端方法，便于单测注入
type EMClient interface {
	GetOrderBook(ctx context.Context, code string) (*em.OrderBookData, error)
	GetKlineHistoryWithKlt(ctx context.Context, code string, limit int, klt int) ([]*em.KlineItem, error)
}

// EMRealtime 可选的实时行情能力
type EMRealtime interface {
	GetRealtimeQuote(ctx context.Context, code string) (*em.RealtimeQuoteData, error)
}
// Market 实现 MarketDataClient
// 说明：
// - K 线优先使用东财 push2his/push2，多主机重试；日线失败时回退网易 CSV
// - 实时报价优先使用东财 push2 字段；失败回退盘口均值
// - 所有调用为实时请求，不引入缓存
type Market struct {
	client EMClient
}

// NewMarket 构造
func NewMarket(c EMClient) *Market {
	return &Market{client: c}
}

// GetOrderBook 使用东财盘口数据
func (m *Market) GetOrderBook(ctx context.Context, code string, depth int) (*provider.OrderBook, error) {
	ob, err := m.client.GetOrderBook(ctx, code)
	if err != nil {
		return nil, err
	}
	// 映射五档
	res := &provider.OrderBook{
		WeiBi:  ob.WeiBi,
		WeiCha: ob.WeiCha,
		UpdatedAt: time.Now(),
	}
	res.Bids[0] = provider.Level{Price: ob.Buy1Price, Volume: int64(ob.Buy1Vol)}
	res.Bids[1] = provider.Level{Price: ob.Buy2Price, Volume: int64(ob.Buy2Vol)}
	res.Bids[2] = provider.Level{Price: ob.Buy3Price, Volume: int64(ob.Buy3Vol)}
	res.Bids[3] = provider.Level{Price: ob.Buy4Price, Volume: int64(ob.Buy4Vol)}
	res.Bids[4] = provider.Level{Price: ob.Buy5Price, Volume: int64(ob.Buy5Vol)}
	res.Asks[0] = provider.Level{Price: ob.Sell1Price, Volume: int64(ob.Sell1Vol)}
	res.Asks[1] = provider.Level{Price: ob.Sell2Price, Volume: int64(ob.Sell2Vol)}
	res.Asks[2] = provider.Level{Price: ob.Sell3Price, Volume: int64(ob.Sell3Vol)}
	res.Asks[3] = provider.Level{Price: ob.Sell4Price, Volume: int64(ob.Sell4Vol)}
	res.Asks[4] = provider.Level{Price: ob.Sell5Price, Volume: int64(ob.Sell5Vol)}
	return res, nil
}

// GetKlines 使用东财 K 线，period 映射 klt
func (m *Market) GetKlines(ctx context.Context, code string, period string, limit int) ([]*provider.KlineItem, error) {
	klt := 101
	switch period {
	case "1m":
		klt = 1
	case "5m":
		klt = 5
	case "15m":
		klt = 15
	case "30m":
		klt = 30
	case "60m":
		klt = 60
	case "day":
		klt = 101
	case "week":
		klt = 102
	case "month":
		klt = 103
	default:
		return nil, fmt.Errorf("不支持的周期: %s", period)
	}
	items, err := m.client.GetKlineHistoryWithKlt(ctx, code, limit, klt)
	if err != nil {
		if period == "day" {
			fb, ferr := netease.GetDailyKline(ctx, code, limit)
			if ferr == nil && len(fb) > 0 {
				var out []*provider.KlineItem
				for _, it := range fb {
					t, _ := time.Parse("2006-01-02", it.Date)
					out = append(out, &provider.KlineItem{
						Time:      t,
						Open:      it.Open,
						Close:     it.Close,
						High:      it.High,
						Low:       it.Low,
						Volume:    it.Volume,
						ChangePct: it.ChangePercent,
					})
				}
				return out, nil
			}
		}
		return nil, err
	}
	var out []*provider.KlineItem
	for _, it := range items {
		// it.Date 格式为 "YYYY-MM-DD HH:MM"
		t, _ := time.Parse("2006-01-02 15:04", it.Date)
		out = append(out, &provider.KlineItem{
			Time:      t,
			Open:      it.Open,
			Close:     it.Close,
			High:      it.High,
			Low:       it.Low,
			Volume:    it.Volume,
			ChangePct: it.ChangePercent,
		})
	}
	return out, nil
}

// GetIntradayQuote 优先使用东财实时行情，失败时回退盘口均值
func (m *Market) GetIntradayQuote(ctx context.Context, code string) (*provider.IntradayQuote, error) {
	if rt, ok := m.client.(EMRealtime); ok {
		if q, err := rt.GetRealtimeQuote(ctx, code); err == nil && q != nil {
		ts := time.Now()
		if q.Timestamp != "" {
			if t, e := time.Parse("2006-01-02 15:04:05", q.Timestamp); e == nil {
				ts = t
			}
		}
		changePct := 0.0
		if q.PrevClose > 0 {
			changePct = (q.Current - q.PrevClose) / q.PrevClose * 100
		}
		return &provider.IntradayQuote{
			Code:      q.Code,
			Name:      q.Name,
			Price:     q.Current,
			Open:      q.Open,
			High:      q.High,
			Low:       q.Low,
			PrevClose: q.PrevClose,
			ChangePct: changePct,
			Volume:    q.Volume,
			Amount:    q.Amount,
			Timestamp: ts,
		}, nil
		}
	}
	// 回退：通过盘口数据估算当前价格（若卖一/买一存在）
	ob, err := m.client.GetOrderBook(ctx, code)
	if err != nil {
		return nil, err
	}
	price := 0.0
	if ob.Buy1Price > 0 && ob.Sell1Price > 0 {
		price = (ob.Buy1Price + ob.Sell1Price) / 2
	} else if ob.Buy1Price > 0 {
		price = ob.Buy1Price
	} else if ob.Sell1Price > 0 {
		price = ob.Sell1Price
	}
	return &provider.IntradayQuote{
		Code:      code,
		Name:      "",
		Price:     price,
		Open:      0,
		High:      0,
		Low:       0,
		PrevClose: 0,
		ChangePct: 0,
		Volume:    0,
		Amount:    0,
		Timestamp: time.Now(),
	}, nil
}

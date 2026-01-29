package eastmoneyimpl_test

import (
	"context"
	"fmt"
	"testing"

	em "stock_assistant/backend/common/eastmoney"
	eimpl "stock_assistant/backend/common/provider/impl/eastmoney"
)

type mockEM struct{}

func (m *mockEM) GetOrderBook(ctx context.Context, code string) (*em.OrderBookData, error) {
	return &em.OrderBookData{
		Buy1Price: 10.0, Buy1Vol: 100,
		Buy2Price: 9.9, Buy2Vol: 200,
		Buy3Price: 9.8, Buy3Vol: 300,
		Buy4Price: 9.7, Buy4Vol: 400,
		Buy5Price: 9.6, Buy5Vol: 500,
		Sell1Price: 10.1, Sell1Vol: 110,
		Sell2Price: 10.2, Sell2Vol: 120,
		Sell3Price: 10.3, Sell3Vol: 130,
		Sell4Price: 10.4, Sell4Vol: 140,
		Sell5Price: 10.5, Sell5Vol: 150,
		WeiBi:  55.5,
		WeiCha: 10,
	}, nil
}

func (m *mockEM) GetKlineHistoryWithKlt(ctx context.Context, code string, limit int, klt int) ([]*em.KlineItem, error) {
	return []*em.KlineItem{
		{Date: "2025-01-02 15:00", Open: 9.5, Close: 10.0, High: 10.2, Low: 9.4, Volume: 12345, ChangePercent: 5.0},
	}, nil
}

func (m *mockEM) GetRealtimeQuote(ctx context.Context, code string) (*em.RealtimeQuoteData, error) {
	return &em.RealtimeQuoteData{
		Code:      code,
		Name:      "测试股",
		Current:   10.2,
		PrevClose: 9.8,
		Open:      9.9,
		High:      10.5,
		Low:       9.7,
		Volume:    54321,
		Amount:    1234567,
		Timestamp: "2025-01-02 14:59:30",
	}, nil
}

func TestGetOrderBook(t *testing.T) {
	market := eimpl.NewMarket(&mockEM{})
	ob, err := market.GetOrderBook(context.Background(), "SH600000", 5)
	if err != nil {
		t.Fatalf("错误: %v", err)
	}
	if ob.Bids[0].Price != 10.0 || ob.Asks[0].Price != 10.1 {
		t.Fatalf("盘口映射错误: buy1=%.2f sell1=%.2f", ob.Bids[0].Price, ob.Asks[0].Price)
	}
	if ob.WeiBi != 55.5 || ob.WeiCha != 10 {
		t.Fatalf("委比/委差错误: %.2f %.2f", ob.WeiBi, ob.WeiCha)
	}
}

func TestGetKlines(t *testing.T) {
	market := eimpl.NewMarket(&mockEM{})
	items, err := market.GetKlines(context.Background(), "SH600000", "day", 1)
	if err != nil {
		t.Fatalf("错误: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("返回数量错误: %d", len(items))
	}
	if items[0].Close != 10.0 || items[0].Volume != 12345 {
		t.Fatalf("K线映射错误: close=%.2f vol=%d", items[0].Close, items[0].Volume)
	}
}

func TestGetIntradayQuote(t *testing.T) {
	market := eimpl.NewMarket(&mockEM{})
	q, err := market.GetIntradayQuote(context.Background(), "SH600000")
	if err != nil {
		t.Fatalf("错误: %v", err)
	}
	// 价格来源实时行情
	if q.Price != 10.2 || q.PrevClose != 9.8 || q.Volume != 54321 {
		t.Fatalf("实时行情映射错误: %+v", q)
	}
}

type mockEMNoRT struct{ mockEM }

func (m *mockEMNoRT) GetRealtimeQuote(ctx context.Context, code string) (*em.RealtimeQuoteData, error) {
	return nil, fmt.Errorf("no rt")
}

func TestGetIntradayQuoteFallback(t *testing.T) {
	market := eimpl.NewMarket(&mockEMNoRT{})
	q, err := market.GetIntradayQuote(context.Background(), "SH600000")
	if err != nil {
		t.Fatalf("错误: %v", err)
	}
	// 回退为买一卖一均值
	if q.Price != 10.05 {
		t.Fatalf("回退报价计算错误: %.2f", q.Price)
	}
}

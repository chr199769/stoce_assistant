package eastmoneyimpl_test

import (
	"context"
	"testing"

	em "stock_assistant/backend/common/eastmoney"
	eimpl "stock_assistant/backend/common/provider/impl/eastmoney"
)

type mockDragon struct{}

func (m *mockDragon) GetDragonTigerHistory(ctx context.Context, code string, limit int) ([]*em.DragonTigerItem, error) {
	return []*em.DragonTigerItem{
		{Date: "2025-01-02", Code: code, Name: "测试股", ClosePrice: 10.0, ChangePercent: 5.0, Reason: "涨停", NetInflow: 1000000},
	}, nil
}

func (m *mockDragon) GetDragonTigerList(ctx context.Context, date string) ([]*em.DragonTigerItem, error) {
	return []*em.DragonTigerItem{
		{Date: date, Code: "SZ000001", Name: "平安银行", ClosePrice: 12.3, ChangePercent: 3.2, Reason: "日涨幅", NetInflow: 500000},
	}, nil
}

func (m *mockDragon) GetDragonTigerSeats(ctx context.Context, code, date string) ([]*em.DragonTigerSeat, []*em.DragonTigerSeat, error) {
	buy := []*em.DragonTigerSeat{{Name: "机构专用", BuyAmt: 300000, SellAmt: 0, NetAmt: 300000}}
	sell := []*em.DragonTigerSeat{{Name: "营业部A", BuyAmt: 0, SellAmt: 200000, NetAmt: -200000}}
	return buy, sell, nil
}

func TestDragonHistory(t *testing.T) {
	cli := eimpl.NewDragonTiger(&mockDragon{})
	items, err := cli.GetHistory(context.Background(), "SZ000001", 1)
	if err != nil {
		t.Fatalf("错误: %v", err)
	}
	if len(items) != 1 || items[0].Code != "SZ000001" || items[0].NetInflow != 1000000 {
		t.Fatalf("历史映射错误: %+v", items)
	}
}

func TestDragonToday(t *testing.T) {
	cli := eimpl.NewDragonTiger(&mockDragon{})
	items, err := cli.GetTodayList(context.Background(), "2025-01-02")
	if err != nil {
		t.Fatalf("错误: %v", err)
	}
	if len(items) != 1 || items[0].Code != "SZ000001" || items[0].Reason == "" {
		t.Fatalf("当日映射错误: %+v", items)
	}
}

func TestDragonSeats(t *testing.T) {
	cli := eimpl.NewDragonTiger(&mockDragon{})
	seats, err := cli.GetSeats(context.Background(), "2025-01-02", "SZ000001")
	if err != nil {
		t.Fatalf("错误: %v", err)
	}
	if len(seats) != 2 {
		t.Fatalf("席位数量错误: %d", len(seats))
	}
}

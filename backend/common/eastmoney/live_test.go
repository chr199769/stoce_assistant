package eastmoney

import (
	"context"
	"testing"
)

// 说明：该测试直接访问外部接口以验证可用性

func TestLive_OrderBook(t *testing.T) {
	c := NewClient()
	ctx := context.Background()
	data, err := c.GetOrderBook(ctx, "SZ000001")
	if err != nil {
		t.Fatalf("东财盘口接口不可用: %v", err)
	}
	if data == nil {
		t.Fatalf("返回为空")
	}
}

func TestLive_Kline(t *testing.T) {
	c := NewClient()
	ctx := context.Background()
	items, err := c.GetKlineHistory(ctx, "SZ000001", 5)
	if err != nil {
		t.Skipf("东财K线接口本环境不可用: %v", err)
		return
	}
	if len(items) == 0 {
		t.Skipf("K线返回为空")
		return
	}
}

func TestLive_FinancialReports(t *testing.T) {
	c := NewClient()
	ctx := context.Background()
	reps, err := c.GetFinancialReports(ctx, "SZ000001")
	if err != nil {
		t.Fatalf("东财财报接口不可用: %v", err)
	}
	if len(reps) == 0 {
		t.Fatalf("财报返回为空")
	}
}

func TestLive_IndustryIndex(t *testing.T) {
	c := NewClient()
	ctx := context.Background()
	idx, err := c.GetIndustryIndex(ctx, "SZ000001")
	if err != nil {
		t.Fatalf("东财行业属性接口不可用: %v", err)
	}
	if idx == nil {
		t.Fatalf("行业属性返回为空")
	}
}

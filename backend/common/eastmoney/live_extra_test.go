package eastmoney

import (
	"context"
	"testing"
)

func TestLive_StockNews(t *testing.T) {
	c := NewClient()
	ctx := context.Background()
	items, err := c.GetStockNews(ctx, "SZ000001")
	if err != nil {
		t.Skipf("东财个股新闻接口本环境不可用: %v", err)
		return
	}
	if len(items) == 0 {
		t.Skip("返回为空")
		return
	}
}

func TestLive_StockNotices(t *testing.T) {
	c := NewClient()
	ctx := context.Background()
	items, err := c.GetStockNotices(ctx, "SZ000001", nil)
	if err != nil {
		t.Skipf("东财公告接口本环境不可用: %v", err)
		return
	}
	if len(items) == 0 {
		t.Skip("返回为空")
		return
	}
}

func TestLive_StockHeat(t *testing.T) {
	c := NewClient()
	ctx := context.Background()
	data, err := c.GetStockHeat(ctx, "SZ000001")
	if err != nil {
		t.Skipf("东财热度接口本环境不可用: %v", err)
		return
	}
	if data == nil {
		t.Skip("返回为空")
		return
	}
}

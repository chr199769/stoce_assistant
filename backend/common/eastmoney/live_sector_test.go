package eastmoney

import (
	"context"
	"testing"
)

func TestLive_SectorRank(t *testing.T) {
	c := NewClient()
	ctx := context.Background()
	items, err := c.GetSectorRank(ctx, "industry", 20)
	if err != nil {
		t.Skipf("东财板块排行接口本环境不可用: %v", err)
		return
	}
	if len(items) == 0 {
		t.Skip("返回为空")
		return
	}
}

func TestLive_SectorList(t *testing.T) {
	c := NewClient()
	ctx := context.Background()
	inds, err := c.GetSectorList(ctx, "industry")
	if err != nil {
		t.Skipf("东财行业列表接口本环境不可用: %v", err)
		return
	}
	concepts, err2 := c.GetSectorList(ctx, "concept")
	if err2 != nil {
		t.Skipf("东财概念列表接口本环境不可用: %v", err2)
		return
	}
	if len(inds) == 0 && len(concepts) == 0 {
		t.Skip("返回为空")
		return
	}
}

func TestLive_AStockList(t *testing.T) {
	c := NewClient()
	ctx := context.Background()
	stocks, err := c.GetAStockList(ctx)
	if err != nil {
		t.Skipf("东财A股清单接口本环境不可用: %v", err)
		return
	}
	if len(stocks) == 0 {
		t.Skip("返回为空")
		return
	}
}

package eastmoneyimpl

import (
	"context"
	"testing"

	em "stock_assistant/backend/common/eastmoney"
)

func TestLive_PopularityAdapter(t *testing.T) {
	c := em.NewClient()
	a := NewPopularity(c)
	ctx := context.Background()
	data, err := a.GetStockHeat(ctx, "SZ000001")
	if err != nil {
		t.Skipf("热度适配接口本环境不可用: %v", err)
		return
	}
	if data == nil {
		t.Skip("返回为空")
		return
	}
}

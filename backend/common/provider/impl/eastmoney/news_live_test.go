package eastmoneyimpl

import (
	"context"
	"testing"

	em "stock_assistant/backend/common/eastmoney"
)

func TestLive_NewsAdapter(t *testing.T) {
	c := em.NewClient()
	a := NewNews(c)
	ctx := context.Background()
	items, err := a.GetStockNews(ctx, "SZ000001", 5)
	if err != nil {
		t.Skipf("新闻适配接口本环境不可用: %v", err)
		return
	}
	if len(items) == 0 {
		t.Skip("返回为空")
		return
	}
}

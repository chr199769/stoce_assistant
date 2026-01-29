package eastmoneyimpl

import (
	"context"
	"testing"

	em "stock_assistant/backend/common/eastmoney"
)

func TestLive_NoticeAdapter(t *testing.T) {
	c := em.NewClient()
	a := NewNotice(c)
	ctx := context.Background()
	items, err := a.GetStockNotices(ctx, "SZ000001", "", 5)
	if err != nil {
		t.Skipf("公告适配接口本环境不可用: %v", err)
		return
	}
	if len(items) == 0 {
		t.Skip("返回为空")
		return
	}
}

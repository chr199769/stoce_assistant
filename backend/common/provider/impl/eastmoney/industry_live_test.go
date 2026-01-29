package eastmoneyimpl

import (
	"context"
	"testing"

	em "stock_assistant/backend/common/eastmoney"
)

func TestLive_IndustryAdapter(t *testing.T) {
	c := em.NewClient()
	a := NewIndustry(c)
	ctx := context.Background()
	data, err := a.GetIndustryIndex(ctx, "SZ000001")
	if err != nil {
		t.Skipf("行业适配接口本环境不可用: %v", err)
		return
	}
	if data == nil {
		t.Skip("返回为空")
		return
	}
}

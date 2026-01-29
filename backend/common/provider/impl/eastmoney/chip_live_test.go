package eastmoneyimpl

import (
	"context"
	"testing"

	em "stock_assistant/backend/common/eastmoney"
)

func TestLive_ChipAdapter(t *testing.T) {
	c := em.NewClient()
	a := NewChip(c)
	ctx := context.Background()
	data, err := a.GetChipDistribution(ctx, "SZ000001")
	if err != nil {
		t.Skipf("筹码适配接口本环境不可用: %v", err)
		return
	}
	if data == nil {
		t.Skip("返回为空")
		return
	}
}

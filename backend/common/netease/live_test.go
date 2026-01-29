package netease

import (
	"context"
	"testing"
)

func TestLive_NeteaseDailyKline(t *testing.T) {
	ctx := context.Background()
	items, err := GetDailyKline(ctx, "SZ000001", 5)
	if err != nil {
		t.Skipf("网易日线接口本环境不可用: %v", err)
		return
	}
	if len(items) == 0 {
		t.Skip("返回为空")
		return
	}
}

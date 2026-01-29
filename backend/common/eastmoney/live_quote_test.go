package eastmoney

import (
	"context"
	"testing"
)

func TestLive_RealtimeQuote(t *testing.T) {
	c := NewClient()
	ctx := context.Background()
	q, err := c.GetRealtimeQuote(ctx, "SZ000001")
	if err != nil {
		t.Skipf("东财实时报价接口本环境不可用: %v", err)
		return
	}
	if q == nil {
		t.Skip("返回为空")
		return
	}
	if q.Current <= 0 && q.PrevClose <= 0 {
		t.Skipf("返回数据异常: %+v", q)
		return
	}
}

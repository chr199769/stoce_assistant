package main

import (
	"context"
	"testing"
	"time"

	stock "stock_assistant/backend/stock_service/kitex_gen/stock"
)

func TestLive_GetRealtime(t *testing.T) {
	svc := NewStockServiceImpl()
	ctx := context.Background()
	req := &stock.GetRealtimeRequest{Code: "sh000001"}
	resp, err := svc.GetRealtime(ctx, req)
	if err != nil {
		t.Skipf("GetRealtime 本环境不可用: %v", err)
		return
	}
	if resp == nil || resp.Stock == nil {
		t.Skip("返回为空")
		return
	}
}

func TestLive_GetMarketSectors(t *testing.T) {
	svc := NewStockServiceImpl()
	ctx := context.Background()
	req := &stock.GetMarketSectorsRequest{Type: "concept", Limit: 10}
	resp, err := svc.GetMarketSectors(ctx, req)
	if err != nil {
		t.Skipf("GetMarketSectors 本环境不可用: %v", err)
		return
	}
	if len(resp.Sectors) == 0 {
		t.Skip("返回为空")
		return
	}
}

func TestLive_GetDragonTigerList(t *testing.T) {
	svc := NewStockServiceImpl()
	ctx := context.Background()
	date := time.Now().Format("2006-01-02")
	req := &stock.GetDragonTigerListRequest{Date: date}
	resp, err := svc.GetDragonTigerList(ctx, req)
	if err != nil {
		t.Skipf("GetDragonTigerList 本环境不可用: %v", err)
		return
	}
	if len(resp.Items) == 0 {
		t.Skip("返回为空")
		return
	}
}

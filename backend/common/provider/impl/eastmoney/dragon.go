package eastmoneyimpl

import (
	"context"
	"time"

	em "stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/common/provider"
)

// EMClientDragon 接口，便于单测替换
type EMClientDragon interface {
	GetDragonTigerHistory(ctx context.Context, code string, limit int) ([]*em.DragonTigerItem, error)
	GetDragonTigerList(ctx context.Context, date string) ([]*em.DragonTigerItem, error)
	GetDragonTigerSeats(ctx context.Context, code, date string) ([]*em.DragonTigerSeat, []*em.DragonTigerSeat, error)
}

// DragonTiger 实现 DragonTigerClient
type DragonTiger struct {
	client EMClientDragon
}

// NewDragonTiger 构造
func NewDragonTiger(c EMClientDragon) *DragonTiger {
	return &DragonTiger{client: c}
}

// GetHistory 映射历史龙虎榜
func (d *DragonTiger) GetHistory(ctx context.Context, code string, limit int) ([]*provider.DragonTigerItem, error) {
	items, err := d.client.GetDragonTigerHistory(ctx, code, limit)
	if err != nil {
		return nil, err
	}
	var out []*provider.DragonTigerItem
	for _, it := range items {
		out = append(out, &provider.DragonTigerItem{
			Date:          it.Date,
			Code:          it.Code,
			Name:          it.Name,
			ClosePrice:    it.ClosePrice,
			ChangePercent: it.ChangePercent,
			Reason:        it.Reason,
			NetInflow:     it.NetInflow,
		})
	}
	return out, nil
}

// GetTodayList 使用东财当日龙虎榜
func (d *DragonTiger) GetTodayList(ctx context.Context, date string) ([]*provider.DragonTigerItem, error) {
	items, err := d.client.GetDragonTigerList(ctx, date)
	if err != nil {
		return nil, err
	}
	var out []*provider.DragonTigerItem
	for _, it := range items {
		out = append(out, &provider.DragonTigerItem{
			Date:          it.Date,
			Code:          it.Code,
			Name:          it.Name,
			ClosePrice:    it.ClosePrice,
			ChangePercent: it.ChangePercent,
			Reason:        it.Reason,
			NetInflow:     it.NetInflow,
		})
	}
	return out, nil
}

// GetSeats 席位明细
func (d *DragonTiger) GetSeats(ctx context.Context, date string, code string) ([]*provider.SeatRecord, error) {
	buy, sell, err := d.client.GetDragonTigerSeats(ctx, date, code)
	if err != nil {
		return nil, err
	}
	var out []*provider.SeatRecord
	for _, it := range buy {
		out = append(out, &provider.SeatRecord{
			SeatName:  it.Name,
			BuyAmount: it.BuyAmt,
			SellAmount: 0,
			NetAmount:  it.NetAmt,
		})
	}
	for _, it := range sell {
		out = append(out, &provider.SeatRecord{
			SeatName:  it.Name,
			BuyAmount: 0,
			SellAmount: it.SellAmt,
			NetAmount:  it.NetAmt,
		})
	}
	_ = time.Now()
	return out, nil
}

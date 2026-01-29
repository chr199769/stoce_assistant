package eastmoneyimpl

import (
	"context"

	em "stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/common/provider"
)

type EMClientPopularity interface {
	GetStockHeat(ctx context.Context, code string) (*em.StockHeatData, error)
}

type Popularity struct {
	client EMClientPopularity
}

func NewPopularity(c EMClientPopularity) *Popularity {
	return &Popularity{client: c}
}

func (p *Popularity) GetStockHeat(ctx context.Context, code string) (*provider.Popularity, error) {
	data, err := p.client.GetStockHeat(ctx, code)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	return &provider.Popularity{
		Rank:      data.Rank,
		HeatScore: data.Heat,
		Source:    "EastMoney",
	}, nil
}

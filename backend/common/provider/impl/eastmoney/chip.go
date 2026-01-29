package eastmoneyimpl

import (
	"context"

	em "stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/common/provider"
)

type EMClientChip interface {
	GetChipDistribution(ctx context.Context, code string) (*em.ChipDistributionData, error)
}

type Chip struct {
	client EMClientChip
}

func NewChip(c EMClientChip) *Chip {
	return &Chip{client: c}
}

func (c *Chip) GetChipDistribution(ctx context.Context, code string) (*provider.ChipDistribution, error) {
	data, err := c.client.GetChipDistribution(ctx, code)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	return &provider.ChipDistribution{
		AverageCost: data.AverageCost,
		WinnerRate:  data.WinnerRate,
		Cost90Low:   data.Cost90Low,
		Cost90High:  data.Cost90High,
	}, nil
}

package eastmoneyimpl

import (
	"context"

	em "stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/common/provider"
)

type EMClientIndustry interface {
	GetIndustryIndex(ctx context.Context, code string) (*em.IndustryIndexData, error)
}

type Industry struct {
	client EMClientIndustry
}

func NewIndustry(c EMClientIndustry) *Industry {
	return &Industry{client: c}
}

func (i *Industry) GetIndustryIndex(ctx context.Context, code string) (*provider.IndustryIndex, error) {
	data, err := i.client.GetIndustryIndex(ctx, code)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	return &provider.IndustryIndex{
		IndustryName: data.IndustryName,
		RegionName:   data.RegionName,
		ConceptNames: data.ConceptNames,
	}, nil
}

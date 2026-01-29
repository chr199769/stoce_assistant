package eastmoneyimpl

import (
	"context"

	em "stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/common/provider"
)

// EMClientSector 适配板块相关能力
type EMClientSector interface {
	GetSectorRank(ctx context.Context, rankType string, limit int) ([]*em.SectorInfo, error)
	GetSectorStocksAll(ctx context.Context, sectorCode string) ([]*em.SectorStockItem, error)
	GetAStockList(ctx context.Context) ([]*em.AStockItem, error)
	GetSectorStocksDetail(ctx context.Context, sectorCode string) ([]*em.SectorStockItem, error)
}

type Sector struct {
	client EMClientSector
}

func NewSector(c EMClientSector) *Sector {
	return &Sector{client: c}
}

func (s *Sector) GetSectorRank(ctx context.Context, kind string) ([]*provider.SectorRankItem, error) {
	items, err := s.client.GetSectorRank(ctx, kind, 50)
	if err != nil {
		return nil, err
	}
	var out []*provider.SectorRankItem
	for _, it := range items {
		out = append(out, &provider.SectorRankItem{
			Code:      it.Code,
			Name:      it.Name,
			ChangePct: it.ChangePercent,
			NetInflow: it.NetInflow,
		})
	}
	return out, nil
}

func (s *Sector) GetSectorStocks(ctx context.Context, sectorCode string) ([]*provider.StockBasic, error) {
	items, err := s.client.GetSectorStocksAll(ctx, sectorCode)
	if err != nil {
		return nil, err
	}
	var out []*provider.StockBasic
	for _, it := range items {
		out = append(out, &provider.StockBasic{
			Code: it.Code,
			Name: it.Name,
		})
	}
	return out, nil
}

func (s *Sector) GetAStockList(ctx context.Context) ([]*provider.StockBasic, error) {
	items, err := s.client.GetAStockList(ctx)
	if err != nil {
		return nil, err
	}
	var out []*provider.StockBasic
	for _, it := range items {
		out = append(out, &provider.StockBasic{
			Code: it.Code,
			Name: it.Name,
		})
	}
	return out, nil
}

func (s *Sector) GetSectorStocksDetail(ctx context.Context, sectorCode string) ([]*provider.SectorStockDetail, error) {
	items, err := s.client.GetSectorStocksDetail(ctx, sectorCode)
	if err != nil {
		return nil, err
	}
	var out []*provider.SectorStockDetail
	for _, it := range items {
		out = append(out, &provider.SectorStockDetail{
			Code:       it.Code,
			Name:       it.Name,
			Price:      it.Price,
			ChangePct:  it.ChangePercent,
			Volume:     it.Volume,
			Amount:     it.Amount,
			MarketCap:  it.MarketCap,
		})
	}
	return out, nil
}
func (s *Sector) GetStockConceptsTHS(ctx context.Context, code string) ([]*provider.ConceptTag, error) {
	return nil, &ProviderError{Source: "eastmoney", Message: "THS 概念不由东财提供"}
}

type ProviderError struct {
	Source  string
	Message string
}

func (e *ProviderError) Error() string {
	return e.Source + ": " + e.Message
}

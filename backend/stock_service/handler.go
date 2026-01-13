package main

import (
	"context"
	"fmt"
	"stock_assistant/backend/stock_service/biz/provider/sina"
	stock "stock_assistant/backend/stock_service/kitex_gen/stock"
)

// StockServiceImpl implements the last service interface defined in the IDL.
type StockServiceImpl struct {
	sinaClient *sina.Client
}

// NewStockServiceImpl creates a new StockServiceImpl
func NewStockServiceImpl() *StockServiceImpl {
	return &StockServiceImpl{
		sinaClient: sina.NewClient(),
	}
}

// GetRealtime implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetRealtime(ctx context.Context, req *stock.GetRealtimeRequest) (resp *stock.GetRealtimeResponse, err error) {
	if req.Code == "" {
		return &stock.GetRealtimeResponse{}, nil
	}

	info, err := s.sinaClient.GetStockInfo(ctx, req.Code)
	if err != nil {
		// Log error and return empty response or specific error code
		// For now, return error
		return nil, err
	}

	return &stock.GetRealtimeResponse{
		Stock: info,
	}, nil
}

// GetMarketSummary implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetMarketSummary(ctx context.Context, req *stock.GetMarketSummaryRequest) (resp *stock.GetMarketSummaryResponse, err error) {
	// Real market indices data
	codes := []string{"sh000001", "sz399001", "sz399006"}
	indices := make([]*stock.MarketIndex, 0)

	for _, code := range codes {
		info, err := s.sinaClient.GetStockInfo(ctx, code)
		if err != nil {
			// Skip or log error? For now, if one fails, we might just continue or return error.
			// Let's try to get as much as possible.
			continue
		}
		indices = append(indices, &stock.MarketIndex{
			Name:          info.Name,
			Value:         info.CurrentPrice,
			Change:        info.CurrentPrice * info.ChangePercent / 100, // Approximate change value
			ChangePercent: info.ChangePercent,
		})
	}

	if len(indices) == 0 {
		return nil, fmt.Errorf("failed to fetch any market indices")
	}

	return &stock.GetMarketSummaryResponse{
		Indices: indices,
	}, nil
}

// GetMarketSectors implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetMarketSectors(ctx context.Context, req *stock.GetMarketSectorsRequest) (resp *stock.GetMarketSectorsResponse, err error) {
	// TODO: Implement real sector data fetching
	// For now return empty to avoid mock data
	return &stock.GetMarketSectorsResponse{
		Sectors: []*stock.SectorInfo{},
	}, nil
}

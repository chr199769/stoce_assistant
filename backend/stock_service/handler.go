package main

import (
	"context"
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
	// Mock market indices data
	indices := []*stock.MarketIndex{
		{
			Name:          "上证指数",
			Value:         3050.12,
			Change:        15.23,
			ChangePercent: 0.52,
		},
		{
			Name:          "深证成指",
			Value:         9850.45,
			Change:        -20.11,
			ChangePercent: -0.21,
		},
		{
			Name:          "创业板指",
			Value:         1950.33,
			Change:        5.67,
			ChangePercent: 0.33,
		},
	}
	return &stock.GetMarketSummaryResponse{
		Indices: indices,
	}, nil
}

// GetMarketSectors implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetMarketSectors(ctx context.Context, req *stock.GetMarketSectorsRequest) (resp *stock.GetMarketSectorsResponse, err error) {
	// Mock market sectors data
	sectors := []*stock.SectorInfo{
		{
			Name:          "人工智能",
			ChangePercent: 2.5,
			Volume:        1000000,
		},
		{
			Name:          "新能源",
			ChangePercent: 1.8,
			Volume:        800000,
		},
		{
			Name:          "医药生物",
			ChangePercent: -0.5,
			Volume:        500000,
		},
	}
	return &stock.GetMarketSectorsResponse{
		Sectors: sectors,
	}, nil
}

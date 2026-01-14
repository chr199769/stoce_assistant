package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"stock_assistant/backend/stock_service/biz/provider/eastmoney"
	"stock_assistant/backend/stock_service/biz/provider/sentiment"
	"stock_assistant/backend/stock_service/biz/provider/sina"
	"stock_assistant/backend/stock_service/dal/redis"
	stock "stock_assistant/backend/stock_service/kitex_gen/stock"
)

// StockServiceImpl implements the last service interface defined in the IDL.
type StockServiceImpl struct {
	sinaClient      *sina.Client
	eastMoneyClient *eastmoney.Client
	sentimentClient *sentiment.Client
}

// NewStockServiceImpl creates a new StockServiceImpl
func NewStockServiceImpl() *StockServiceImpl {
	return &StockServiceImpl{
		sinaClient:      sina.NewClient(),
		eastMoneyClient: eastmoney.NewClient(),
		sentimentClient: sentiment.NewClient(),
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

// GetFinancialReport implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetFinancialReport(ctx context.Context, req *stock.GetFinancialReportRequest) (resp *stock.GetFinancialReportResponse, err error) {
	if req.Code == "" {
		return &stock.GetFinancialReportResponse{}, nil
	}

	reports, err := s.eastMoneyClient.GetFinancialReports(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	return &stock.GetFinancialReportResponse{
		Reports: reports,
	}, nil
}

// GetMarketSectors implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetMarketSectors(ctx context.Context, req *stock.GetMarketSectorsRequest) (resp *stock.GetMarketSectorsResponse, err error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}

	// Default type
	rankType := req.Type
	if rankType == "" {
		rankType = "concept"
	}

	// Try Redis Cache first
	cacheKey := fmt.Sprintf("market:sector:rank:%s", rankType)
	if cached, err := redis.Get(ctx, cacheKey); err == nil && cached != "" {
		var thriftSectors []*stock.SectorInfo
		if err := json.Unmarshal([]byte(cached), &thriftSectors); err == nil {
			// If cached sectors found, check limit
			if len(thriftSectors) > limit {
				thriftSectors = thriftSectors[:limit]
			}
			return &stock.GetMarketSectorsResponse{Sectors: thriftSectors}, nil
		}
	}

	sectors, err := s.eastMoneyClient.GetSectorRank(ctx, rankType, limit)
	if err != nil {
		return nil, err
	}

	// Convert to thrift struct
	var thriftSectors []*stock.SectorInfo
	for _, sec := range sectors {
		thriftSectors = append(thriftSectors, &stock.SectorInfo{
			Code:          sec.Code,
			Name:          sec.Name,
			ChangePercent: sec.ChangePercent,
			NetInflow:     sec.NetInflow,
			TopStockName:  sec.TopStockName,
			TopStockCode:  sec.TopStockCode,
			Type:          rankType,
		})
	}

	// Set Redis Cache (TTL 60s)
	if bytes, err := json.Marshal(thriftSectors); err == nil {
		_ = redis.Set(ctx, cacheKey, string(bytes), 60*time.Second)
	}

	return &stock.GetMarketSectorsResponse{Sectors: thriftSectors}, nil
}

// GetLimitUpPool implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetLimitUpPool(ctx context.Context, req *stock.GetLimitUpPoolRequest) (resp *stock.GetLimitUpPoolResponse, err error) {
	// Try Redis Cache first (only if date is not specified or is today)
	// For simplicity, we only cache the "current" pool
	cacheKey := "market:limit_up:pool"
	
	if cached, err := redis.Get(ctx, cacheKey); err == nil && cached != "" {
		var thriftStocks []*stock.LimitUpStock
		if err := json.Unmarshal([]byte(cached), &thriftStocks); err == nil {
			return &stock.GetLimitUpPoolResponse{Stocks: thriftStocks}, nil
		}
	}

	// Note: req.Date is currently ignored by the simple implementation, but could be passed if upgraded.
	pool, err := s.sentimentClient.GetLimitUpPool(ctx)
	if err != nil {
		return nil, err
	}

	var thriftStocks []*stock.LimitUpStock
	for _, item := range pool {
		thriftStocks = append(thriftStocks, &stock.LimitUpStock{
			Code:          item.Code,
			Name:          item.Name,
			Price:         item.Price,
			ChangePercent: item.ChangePercent,
			LimitUpType:   item.LimitUpType,
			Reason:        item.Reason,
			IsBroken:      item.IsBroken,
		})
	}

	// Set Redis Cache (TTL 30s)
	if len(thriftStocks) > 0 {
		if bytes, err := json.Marshal(thriftStocks); err == nil {
			_ = redis.Set(ctx, cacheKey, string(bytes), 30*time.Second)
		}
	}

	return &stock.GetLimitUpPoolResponse{Stocks: thriftStocks}, nil
}

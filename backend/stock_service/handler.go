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
	"stock_assistant/backend/stock_service/dal/model"
	"stock_assistant/backend/stock_service/dal/mysql"
	stock "stock_assistant/backend/stock_service/kitex_gen/stock"
	"sort"
	"strings"
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

// GetSectorStocks implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetSectorStocks(ctx context.Context, req *stock.GetSectorStocksRequest) (resp *stock.GetSectorStocksResponse, err error) {
	if req.SectorCode == "" {
		return &stock.GetSectorStocksResponse{}, nil
	}

	// Call client
	rawStocks, err := s.eastMoneyClient.GetSectorStocksRaw(ctx, req.SectorCode)
	if err != nil {
		return nil, err
	}

	var list []*stock.SectorStockItem
	for _, item := range rawStocks {
		list = append(list, &stock.SectorStockItem{
			Code:          item.Code,
			Name:          item.Name,
			Price:         item.Price,
			ChangePercent: item.ChangePercent,
			Volume:        item.Volume,
			Amount:        item.Amount,
			MarketCap:     item.MarketCap,
		})
	}

	return &stock.GetSectorStocksResponse{Stocks: list}, nil
}

// GetDragonTigerList implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetDragonTigerList(ctx context.Context, req *stock.GetDragonTigerListRequest) (resp *stock.GetDragonTigerListResponse, err error) {
	date := req.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	items, err := s.eastMoneyClient.GetDragonTigerList(ctx, date)
	if err != nil {
		return nil, err
	}

	// Sort by Net Inflow Desc
	sort.Slice(items, func(i, j int) bool {
		return items[i].NetInflow > items[j].NetInflow
	})

	// Seat Mapping Map
	seatMap := map[string]string{
		"华泰证券股份有限公司北京雍和宫证券营业部":      "赵老哥",
		"国泰君安证券股份有限公司上海江苏路证券营业部":     "章盟主",
		"中国银河证券股份有限公司北京绍兴路证券营业部":      "赵老哥",
		"东方财富证券股份有限公司拉萨团结路第二证券营业部": "拉萨天团",
		"东方财富证券股份有限公司拉萨团结路第一证券营业部": "拉萨天团",
		"东方财富证券股份有限公司拉萨东环路第二证券营业部":   "拉萨天团",
		"东方财富证券股份有限公司拉萨东环路第一证券营业部":   "拉萨天团",
		"招商证券股份有限公司深圳益田路免税商务大厦证券营业部": "益田路",
		"中信证券股份有限公司上海溧阳路证券营业部":       "孙哥",
	}

	var thriftItems []*stock.DragonTigerItem

	// Limit detail fetching to Top 10 to avoid timeout
	for i, item := range items {
		tItem := &stock.DragonTigerItem{
			Code:          item.Code,
			Name:          item.Name,
			ClosePrice:    item.ClosePrice,
			ChangePercent: item.ChangePercent,
			Reason:        item.Reason,
			NetInflow:     item.NetInflow,
			BuySeats:      []*stock.DragonTigerSeat{},
			SellSeats:     []*stock.DragonTigerSeat{},
		}

		if i < 5 { // Only fetch seats for top 5
			buySeats, sellSeats, err := s.eastMoneyClient.GetDragonTigerSeats(ctx, item.Code, date)
			if err == nil {
				tItem.BuySeats = convertSeats(buySeats, seatMap)
				tItem.SellSeats = convertSeats(sellSeats, seatMap)
			}
		}
		thriftItems = append(thriftItems, tItem)
	}

	return &stock.GetDragonTigerListResponse{Items: thriftItems}, nil
}

func convertSeats(seats []*eastmoney.DragonTigerSeat, m map[string]string) []*stock.DragonTigerSeat {
	var res []*stock.DragonTigerSeat
	for _, s := range seats {
		tags := []string{}
		if t, ok := m[s.Name]; ok {
			tags = append(tags, t)
		}
		// Add other simple checks
		if strings.Contains(s.Name, "拉萨") && len(tags) == 0 {
			tags = append(tags, "拉萨天团")
		}
		if strings.Contains(s.Name, "机构专用") {
			tags = append(tags, "机构")
		}
		if strings.Contains(s.Name, "沪股通") || strings.Contains(s.Name, "深股通") {
			tags = append(tags, "北向资金")
		}

		res = append(res, &stock.DragonTigerSeat{
			Name:    s.Name,
			BuyAmt:  s.BuyAmt,
			SellAmt: s.SellAmt,
			NetAmt:  s.NetAmt,
			Tags:    tags,
		})
	}
	return res
}

// GetOrCreateUser implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetOrCreateUser(ctx context.Context, req *stock.GetOrCreateUserRequest) (resp *stock.GetOrCreateUserResponse, err error) {
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}

	if mysql.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var user model.User
	// Find or create
	// GORM FirstOrCreate: finds by unique constraint or primary key, or creates.
	// Since Username is uniqueIndex, we search by Username.
	err = mysql.DB.Where(model.User{Username: req.Username}).FirstOrCreate(&user).Error
	if err != nil {
		return nil, err
	}

	return &stock.GetOrCreateUserResponse{
		User: &stock.User{
			Id:        fmt.Sprintf("%d", user.ID), // Convert uint to string
			Username:  user.Username,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// AddWatchlist implements the StockServiceImpl interface.
func (s *StockServiceImpl) AddWatchlist(ctx context.Context, req *stock.AddWatchlistRequest) (resp *stock.AddWatchlistResponse, err error) {
	if mysql.DB == nil {
		return &stock.AddWatchlistResponse{Success: false}, nil
	}

	// Normalize stock code
	code := strings.TrimSpace(req.StockCode)
	if len(code) == 6 {
		if strings.HasPrefix(code, "6") {
			code = "sh" + code
		} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
			code = "sz" + code
		}
	}

	item := model.UserWatchlist{
		UserID:    req.UserId,
		StockCode: code,
		Tags:      "[]", // Default empty JSON array
	}
	// Check if exists
	var count int64
	mysql.DB.Model(&model.UserWatchlist{}).Where("user_id = ? AND stock_code = ?", req.UserId, code).Count(&count)
	if count > 0 {
		return &stock.AddWatchlistResponse{Success: true}, nil
	}

	if err := mysql.DB.Create(&item).Error; err != nil {
		return &stock.AddWatchlistResponse{Success: false}, nil
	}
	return &stock.AddWatchlistResponse{Success: true}, nil
}

// GetWatchlist implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetWatchlist(ctx context.Context, req *stock.GetWatchlistRequest) (resp *stock.GetWatchlistResponse, err error) {
	if mysql.DB == nil {
		return &stock.GetWatchlistResponse{}, nil
	}

	var items []model.UserWatchlist
	if err := mysql.DB.Where("user_id = ?", req.UserId).Find(&items).Error; err != nil {
		return nil, err
	}

	var thriftItems []*stock.WatchlistItem
	for _, item := range items {
		// Simple JSON parse for tags (not robust but sufficient for now)
		// Tags is stored as string "[]" or `["A","B"]`
		tags := []string{}
		if len(item.Tags) > 2 {
			// Strip brackets and split
			content := item.Tags[1 : len(item.Tags)-1]
			if content != "" {
				// Split by comma
				parts := strings.Split(content, ",")
				for _, p := range parts {
					tags = append(tags, strings.Trim(strings.TrimSpace(p), "\""))
				}
			}
		}

		thriftItems = append(thriftItems, &stock.WatchlistItem{
			StockCode: item.StockCode,
			Tags:      tags,
			AddedAt:   item.CreatedAt.Format(time.RFC3339),
		})
	}
	return &stock.GetWatchlistResponse{Items: thriftItems}, nil
}

// RemoveWatchlist implements the StockServiceImpl interface.
func (s *StockServiceImpl) RemoveWatchlist(ctx context.Context, req *stock.RemoveWatchlistRequest) (resp *stock.RemoveWatchlistResponse, err error) {
	if mysql.DB == nil {
		return &stock.RemoveWatchlistResponse{Success: false}, nil
	}

	// Normalize stock code
	code := strings.TrimSpace(req.StockCode)
	if len(code) == 6 {
		if strings.HasPrefix(code, "6") {
			code = "sh" + code
		} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
			code = "sz" + code
		}
	}

	if err := mysql.DB.Where("user_id = ? AND stock_code = ?", req.UserId, code).Delete(&model.UserWatchlist{}).Error; err != nil {
		return &stock.RemoveWatchlistResponse{Success: false}, nil
	}
	return &stock.RemoveWatchlistResponse{Success: true}, nil
}

// SaveIntradaySignal implements the StockServiceImpl interface.
func (s *StockServiceImpl) SaveIntradaySignal(ctx context.Context, req *stock.SaveIntradaySignalRequest) (resp *stock.SaveIntradaySignalResponse, err error) {
	if mysql.DB == nil {
		return &stock.SaveIntradaySignalResponse{Success: false}, nil
	}
	if req.Signal == nil {
		return &stock.SaveIntradaySignalResponse{Success: false}, nil
	}

	sig := model.IntradaySignal{
		StockCode:   req.Signal.StockCode,
		SignalType:  req.Signal.SignalType,
		Score:       req.Signal.Score,
		Description: req.Signal.Description,
		TriggerTime: time.Now(),
		TraceID:     req.Signal.TraceId,
	}
	if err := mysql.DB.Create(&sig).Error; err != nil {
		return &stock.SaveIntradaySignalResponse{Success: false}, nil
	}
	return &stock.SaveIntradaySignalResponse{Success: true}, nil
}

// GetIntradaySignals implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetIntradaySignals(ctx context.Context, req *stock.GetIntradaySignalsRequest) (resp *stock.GetIntradaySignalsResponse, err error) {
	if mysql.DB == nil {
		return &stock.GetIntradaySignalsResponse{}, nil
	}

	var signals []model.IntradaySignal
	// Default to last 50 signals
	// If date is provided, we could filter by date, but keeping it simple for now
	if err := mysql.DB.Order("trigger_time desc").Limit(50).Find(&signals).Error; err != nil {
		return nil, err
	}

	var thriftSignals []*stock.IntradaySignal
	for _, s := range signals {
		thriftSignals = append(thriftSignals, &stock.IntradaySignal{
			StockCode:   s.StockCode,
			SignalType:  s.SignalType,
			Score:       s.Score,
			Description: s.Description,
			TriggerTime: s.TriggerTime.Format(time.RFC3339),
			TraceId:     s.TraceID,
		})
	}
	return &stock.GetIntradaySignalsResponse{Signals: thriftSignals}, nil
}

// GetHistoricalKline implements the StockServiceImpl interface.
func (s *StockServiceImpl) GetHistoricalKline(ctx context.Context, req *stock.GetHistoricalKlineRequest) (resp *stock.GetHistoricalKlineResponse, err error) {
	// Use EastMoney Client
	klines, err := s.eastMoneyClient.GetKlineHistory(ctx, req.StockCode, int(req.Days))
	if err != nil {
		return nil, err
	}

	var thriftKlines []*stock.Kline
	for _, k := range klines {
		thriftKlines = append(thriftKlines, &stock.Kline{
			Date:   k.Date,
			Open:   k.Open,
			Close:  k.Close,
			High:   k.High,
			Low:    k.Low,
			Volume: k.Volume,
		})
	}

	return &stock.GetHistoricalKlineResponse{Klines: thriftKlines}, nil
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"sort"
	eastmoney "stock_assistant/backend/common/eastmoney"
	eimpl "stock_assistant/backend/common/provider/impl/eastmoney"
	"stock_assistant/backend/common/provider"
	"stock_assistant/backend/stock_service/biz/provider/sentiment"
	"stock_assistant/backend/stock_service/biz/provider/sina"
	"stock_assistant/backend/stock_service/dal/model"
	"stock_assistant/backend/stock_service/dal/mysql"
	"stock_assistant/backend/stock_service/dal/redis"
	stock "stock_assistant/backend/stock_service/kitex_gen/stock"
	"strings"
)

// StockServiceImpl 实现 IDL 定义的服务接口
type StockServiceImpl struct {
	sinaClient      *sina.Client
	eastMoneyClient *eastmoney.Client
	sentimentClient *sentiment.Client
	marketClient    provider.MarketDataClient
	sectorClient    provider.SectorClient
	dragonClient    provider.DragonTigerClient
}

// NewStockServiceImpl 创建新的 StockServiceImpl
func NewStockServiceImpl() *StockServiceImpl {
	emc := eastmoney.NewClient()
	return &StockServiceImpl{
		sinaClient:      sina.NewClient(),
		eastMoneyClient: emc,
		sentimentClient: sentiment.NewClient(),
		marketClient:    eimpl.NewMarket(emc),
		sectorClient:    eimpl.NewSector(emc),
		dragonClient:    eimpl.NewDragonTiger(emc),
	}
}

// GetRealtime 实时行情
// 约束：
// - 入参 code 使用 "sh/sz+6位代码"（例：sh600519）
// - 直接调用新浪行情，无缓存；失败原样向上抛出
func (s *StockServiceImpl) GetRealtime(ctx context.Context, req *stock.GetRealtimeRequest) (resp *stock.GetRealtimeResponse, err error) {
	if req.Code == "" {
		return &stock.GetRealtimeResponse{}, nil
	}

	// 优先使用统一 Provider，失败回退新浪
	code := normalizeCode(req.Code)
	if s.marketClient != nil {
		if q, e := s.marketClient.GetIntradayQuote(ctx, code); e == nil && q != nil && q.Price > 0 {
			return &stock.GetRealtimeResponse{
				Stock: &stock.StockInfo{
					Code:          q.Code,
					Name:          q.Name,
					CurrentPrice:  q.Price,
					ChangePercent: q.ChangePct,
					Volume:        q.Volume,
					Timestamp:     q.Timestamp.Format("2006-01-02 15:04:05"),
				},
			}, nil
		}
	}
	info, err := s.sinaClient.GetStockInfo(ctx, code)
	if err != nil {
		return nil, err
	}

	return &stock.GetRealtimeResponse{
		Stock: info,
	}, nil
}

// GetFinancialReport 财报摘要
// 约束：
// - 入参 code 使用 "SH/SZ+6位代码" 或 6 位代码（内部处理）
// - 最近 N 期季度数据，单位与比例统一为元/百分比
func (s *StockServiceImpl) GetFinancialReport(ctx context.Context, req *stock.GetFinancialReportRequest) (resp *stock.GetFinancialReportResponse, err error) {
	if req.Code == "" {
		return &stock.GetFinancialReportResponse{}, nil
	}

	reports, err := s.eastMoneyClient.GetFinancialReports(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	var thriftReports []*stock.FinancialData
	for _, r := range reports {
		thriftReports = append(thriftReports, &stock.FinancialData{
			ReportDate:   r.ReportDate,
			TotalRevenue: r.TotalRevenue,
			NetProfit:    r.NetProfit,
			Eps:          r.Eps,
			RevenueYoy:   r.RevenueYoy,
			ProfitYoy:    r.ProfitYoy,
		})
	}

	return &stock.GetFinancialReportResponse{
		Reports: thriftReports,
	}, nil
}

// GetMarketSectors 板块排行
// 约束：
// - req.Type 支持 "industry"/"concept"，默认 concept
// - 返回前 limit 个；引入 60s 轻缓存保障稳定性
func (s *StockServiceImpl) GetMarketSectors(ctx context.Context, req *stock.GetMarketSectorsRequest) (resp *stock.GetMarketSectorsResponse, err error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}

	// 默认类型
	rankType := req.Type
	if rankType == "" {
		rankType = "concept"
	}

	// 优先尝试 Redis 缓存
	cacheKey := fmt.Sprintf("market:sector:rank:%s", rankType)
	if cached, err := redis.Get(ctx, cacheKey); err == nil && cached != "" {
		var thriftSectors []*stock.SectorInfo
		if err := json.Unmarshal([]byte(cached), &thriftSectors); err == nil {
			// 如果找到缓存，检查 limit
			if len(thriftSectors) > limit {
				thriftSectors = thriftSectors[:limit]
			}
			return &stock.GetMarketSectorsResponse{Sectors: thriftSectors}, nil
		}
	}

	var sectors []*eastmoney.SectorInfo
	if s.sectorClient != nil {
		if items, e := s.sectorClient.GetSectorRank(ctx, rankType); e == nil && len(items) > 0 {
			for _, it := range items {
				sectors = append(sectors, &eastmoney.SectorInfo{
					Code:          it.Code,
					Name:          it.Name,
					ChangePercent: it.ChangePct,
					NetInflow:     it.NetInflow,
				})
			}
		}
	}
	if len(sectors) == 0 {
		sectors, err = s.eastMoneyClient.GetSectorRank(ctx, rankType, limit)
	}
	if err != nil {
		return nil, err
	}

	// 转换为 thrift 结构
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

	// 设置 Redis 缓存 (TTL 60s)
	if bytes, err := json.Marshal(thriftSectors); err == nil {
		_ = redis.Set(ctx, cacheKey, string(bytes), 60*time.Second)
	}

	return &stock.GetMarketSectorsResponse{Sectors: thriftSectors}, nil
}

// GetLimitUpPool 涨停池（情绪）
// 约束：
// - 依赖非官方接口，可能不稳定；为空或异常时返回空列表
// - 缓存 30s，避免频繁请求
func (s *StockServiceImpl) GetLimitUpPool(ctx context.Context, req *stock.GetLimitUpPoolRequest) (resp *stock.GetLimitUpPoolResponse, err error) {
	// 优先尝试 Redis 缓存
	// 为简单起见，我们只缓存“当前”池
	cacheKey := "market:limit_up:pool"

	if cached, err := redis.Get(ctx, cacheKey); err == nil && cached != "" {
		var thriftStocks []*stock.LimitUpStock
		if err := json.Unmarshal([]byte(cached), &thriftStocks); err == nil {
			return &stock.GetLimitUpPoolResponse{Stocks: thriftStocks}, nil
		}
	}

	// 注意: req.Date 目前被简单实现忽略，如果升级可以传递。
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

	// 设置 Redis 缓存 (TTL 30s)
	if len(thriftStocks) > 0 {
		if bytes, err := json.Marshal(thriftStocks); err == nil {
			_ = redis.Set(ctx, cacheKey, string(bytes), 30*time.Second)
		}
	}

	return &stock.GetLimitUpPoolResponse{Stocks: thriftStocks}, nil
}

// GetSectorStocks 板块包含股票列表
// 约束：
// - req.SectorCode 为东财板块代码（b:BKxxx）
// - 返回字段统一语义：价格/涨跌/成交量/成交额/市值
func (s *StockServiceImpl) GetSectorStocks(ctx context.Context, req *stock.GetSectorStocksRequest) (resp *stock.GetSectorStocksResponse, err error) {
	if req.SectorCode == "" {
		return &stock.GetSectorStocksResponse{}, nil
	}

	var list []*stock.SectorStockItem
	if s.sectorClient != nil {
		if items, e := s.sectorClient.GetSectorStocksDetail(ctx, req.SectorCode); e == nil && len(items) > 0 {
			for _, it := range items {
				list = append(list, &stock.SectorStockItem{
					Code:          it.Code,
					Name:          it.Name,
					Price:         it.Price,
					ChangePercent: it.ChangePct,
					Volume:        it.Volume,
					Amount:        it.Amount,
					MarketCap:     it.MarketCap,
				})
			}
			return &stock.GetSectorStocksResponse{Stocks: list}, nil
		}
	}
	// 回退到东财客户端
	rawStocks, err := s.eastMoneyClient.GetSectorStocksDetail(ctx, req.SectorCode)
	if err != nil {
		return nil, err
	}

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

// GetDragonTigerList 龙虎榜
// 约束：
// - req.Date 格式 "YYYY-MM-DD"，为空时默认当天
// - 席位明细仅获取前5，避免超时；席位标签做简单映射
func (s *StockServiceImpl) GetDragonTigerList(ctx context.Context, req *stock.GetDragonTigerListRequest) (resp *stock.GetDragonTigerListResponse, err error) {
	date := req.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	var items []*eastmoney.DragonTigerItem
	if s.dragonClient != nil {
		if list, e := s.dragonClient.GetTodayList(ctx, date); e == nil && len(list) > 0 {
			for _, it := range list {
				items = append(items, &eastmoney.DragonTigerItem{
					Code:          it.Code,
					Name:          it.Name,
					ClosePrice:    it.ClosePrice,
					ChangePercent: it.ChangePercent,
					Reason:        it.Reason,
					NetInflow:     it.NetInflow,
				})
			}
		}
	}
	if len(items) == 0 {
		items, err = s.eastMoneyClient.GetDragonTigerList(ctx, date)
	}
	if err != nil {
		return nil, err
	}

	// 按净流入降序排序
	sort.Slice(items, func(i, j int) bool {
		return items[i].NetInflow > items[j].NetInflow
	})

	// 席位映射表
	seatMap := map[string]string{
		"华泰证券股份有限公司北京雍和宫证券营业部":       "赵老哥",
		"国泰君安证券股份有限公司上海江苏路证券营业部":     "章盟主",
		"中国银河证券股份有限公司北京绍兴路证券营业部":     "赵老哥",
		"东方财富证券股份有限公司拉萨团结路第二证券营业部":   "拉萨天团",
		"东方财富证券股份有限公司拉萨团结路第一证券营业部":   "拉萨天团",
		"东方财富证券股份有限公司拉萨东环路第二证券营业部":   "拉萨天团",
		"东方财富证券股份有限公司拉萨东环路第一证券营业部":   "拉萨天团",
		"招商证券股份有限公司深圳益田路免税商务大厦证券营业部": "益田路",
		"中信证券股份有限公司上海溧阳路证券营业部":       "孙哥",
	}

	var thriftItems []*stock.DragonTigerItem

	// 限制详情获取为前10以避免超时
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

		if i < 5 { // 仅获取前5的席位
			if s.dragonClient != nil {
				if seats, e := s.dragonClient.GetSeats(ctx, date, item.Code); e == nil && len(seats) > 0 {
					var buySeatsEM []*eastmoney.DragonTigerSeat
					var sellSeatsEM []*eastmoney.DragonTigerSeat
					for _, sr := range seats {
						if sr.NetAmount >= 0 {
							buySeatsEM = append(buySeatsEM, &eastmoney.DragonTigerSeat{
								Name:   sr.SeatName,
								BuyAmt: sr.BuyAmount,
								SellAmt: sr.SellAmount,
								NetAmt: sr.NetAmount,
							})
						} else {
							sellSeatsEM = append(sellSeatsEM, &eastmoney.DragonTigerSeat{
								Name:   sr.SeatName,
								BuyAmt: sr.BuyAmount,
								SellAmt: sr.SellAmount,
								NetAmt: sr.NetAmount,
							})
						}
					}
					tItem.BuySeats = convertSeats(buySeatsEM, seatMap)
					tItem.SellSeats = convertSeats(sellSeatsEM, seatMap)
				}
			} else {
				buySeats, sellSeats, err := s.eastMoneyClient.GetDragonTigerSeats(ctx, item.Code, date)
				if err == nil {
					tItem.BuySeats = convertSeats(buySeats, seatMap)
					tItem.SellSeats = convertSeats(sellSeats, seatMap)
				}
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
		// 添加其他简单检查
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

func normalizeCode(in string) string {
	s := strings.ToUpper(strings.TrimSpace(in))
	if strings.HasPrefix(s, "SH") || strings.HasPrefix(s, "SZ") {
		return s
	}
	if len(s) == 6 && s[0] == '6' {
		return "SH" + s
	}
	if len(s) == 6 {
		return "SZ" + s
	}
	return s
}

// GetOrCreateUser 实现 StockServiceImpl 接口
func (s *StockServiceImpl) GetOrCreateUser(ctx context.Context, req *stock.GetOrCreateUserRequest) (resp *stock.GetOrCreateUserResponse, err error) {
	if req.Username == "" {
		return nil, fmt.Errorf("用户名是必须的")
	}

	if mysql.DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var user model.User
	// 查找或创建
	// GORM FirstOrCreate: 按唯一约束或主键查找，否则创建。
	// 由于 Username 是 uniqueIndex，我们按 Username 搜索。
	err = mysql.DB.Where(model.User{Username: req.Username}).FirstOrCreate(&user).Error
	if err != nil {
		return nil, err
	}

	return &stock.GetOrCreateUserResponse{
		User: &stock.User{
			Id:        fmt.Sprintf("%d", user.ID), // 转换 uint 为 string
			Username:  user.Username,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// AddWatchlist 实现 StockServiceImpl 接口
func (s *StockServiceImpl) AddWatchlist(ctx context.Context, req *stock.AddWatchlistRequest) (resp *stock.AddWatchlistResponse, err error) {
	if mysql.DB == nil {
		return &stock.AddWatchlistResponse{Success: false}, nil
	}

	// 规范化股票代码
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
		Tags:      "[]", // 默认空 JSON 数组
	}
	// 检查是否存在
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

// GetWatchlist 实现 StockServiceImpl 接口
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
		// 简单 JSON 解析 tags (目前足够用)
		// Tags 存储为字符串 "[]" 或 `["A","B"]`
		tags := []string{}
		if len(item.Tags) > 2 {
			// 去除括号并分割
			content := item.Tags[1 : len(item.Tags)-1]
			if content != "" {
				// 按逗号分割
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

// RemoveWatchlist 实现 StockServiceImpl 接口
func (s *StockServiceImpl) RemoveWatchlist(ctx context.Context, req *stock.RemoveWatchlistRequest) (resp *stock.RemoveWatchlistResponse, err error) {
	if mysql.DB == nil {
		return &stock.RemoveWatchlistResponse{Success: false}, nil
	}

	// 规范化股票代码
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

// SaveIntradaySignal 实现 StockServiceImpl 接口
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

// GetIntradaySignals 实现 StockServiceImpl 接口
func (s *StockServiceImpl) GetIntradaySignals(ctx context.Context, req *stock.GetIntradaySignalsRequest) (resp *stock.GetIntradaySignalsResponse, err error) {
	if mysql.DB == nil {
		return &stock.GetIntradaySignalsResponse{}, nil
	}

	var signals []model.IntradaySignal
	// 默认最近 50 条信号
	// 如果提供了日期，我们可以按日期过滤，但目前保持简单
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

// GetHistoricalKline 实现 StockServiceImpl 接口
func (s *StockServiceImpl) GetHistoricalKline(ctx context.Context, req *stock.GetHistoricalKlineRequest) (resp *stock.GetHistoricalKlineResponse, err error) {
	// 使用 EastMoney 客户端
	klines, err := s.eastMoneyClient.GetKlineHistory(ctx, req.StockCode, int(req.Days))
	if err != nil {
		log.Printf("获取K线失败: 股票=%s 天数=%d 错误=%v", req.StockCode, req.Days, err)
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

// SavePrediction 实现 StockServiceImpl 接口
func (s *StockServiceImpl) SavePrediction(ctx context.Context, req *stock.SavePredictionRequest) (resp *stock.SavePredictionResponse, err error) {
	if mysql.DB == nil {
		return &stock.SavePredictionResponse{Success: false}, nil
	}
	if req.Record == nil {
		return &stock.SavePredictionResponse{Success: false}, nil
	}

	predDate, _ := time.Parse("2006-01-02 15:04:05", req.Record.PredictionDate)
	if predDate.IsZero() {
		predDate = time.Now()
	}

	record := model.PredictionRecord{
		ID:                req.Record.Id,
		StockCode:         req.Record.StockCode,
		PredictionDate:    predDate,
		Content:           req.Record.Content,
		Confidence:        req.Record.Confidence,
		Trend:             req.Record.Trend,
		TargetPrice:       req.Record.TargetPrice,
		StopLossPrice:     req.Record.StopLossPrice,
		PolicyImpactScope: req.Record.PolicyImpactScope,
		PredictedChange:   req.Record.PredictedChange,
		TraceID:           req.Record.TraceId,
	}

	if err := mysql.DB.Create(&record).Error; err != nil {
		return &stock.SavePredictionResponse{Success: false}, nil
	}

	// 同时创建初始评估记录
	eval := model.EvaluationRecord{
		ID:             fmt.Sprintf("eval-%s", record.ID), // 简单 ID 生成
		PredictionID:   record.ID,
		StockCode:      record.StockCode,
		PredictionDate: record.PredictionDate,
		Status:         "pending",
	}

	// 获取当前价格作为初始价格
	info, err := s.sinaClient.GetStockInfo(ctx, record.StockCode)
	if err == nil {
		eval.InitialPrice = info.CurrentPrice
		eval.StockName = info.Name
	}

	mysql.DB.Create(&eval)

	return &stock.SavePredictionResponse{Success: true}, nil
}

// GetEvaluations 实现 StockServiceImpl 接口
func (s *StockServiceImpl) GetEvaluations(ctx context.Context, req *stock.GetEvaluationsRequest) (resp *stock.GetEvaluationsResponse, err error) {
	if mysql.DB == nil {
		return &stock.GetEvaluationsResponse{}, nil
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	offset := int(req.Offset)

	var evals []model.EvaluationRecord
	query := mysql.DB.Model(&model.EvaluationRecord{}).Order("prediction_date desc").Limit(limit).Offset(offset)

	if req.StockCode != "" {
		query = query.Where("stock_code = ?", req.StockCode)
	}

	if err := query.Find(&evals).Error; err != nil {
		return nil, err
	}

	var thriftEvals []*stock.EvaluationRecord
	for _, e := range evals {
		thriftEvals = append(thriftEvals, &stock.EvaluationRecord{
			Id:             e.ID,
			PredictionId:   e.PredictionID,
			StockCode:      e.StockCode,
			PredictionDate: e.PredictionDate.Format("2006-01-02 15:04:05"),
			InitialPrice:   e.InitialPrice,
			Price_1d:       e.Price1D,
			Price_2d:       e.Price2D,
			Price_3d:       e.Price3D,
			Score:          e.Score,
			Status:         e.Status,
		})
	}

	return &stock.GetEvaluationsResponse{Evaluations: thriftEvals}, nil
}

// DeleteEvaluation 实现 StockServiceImpl 接口
func (s *StockServiceImpl) DeleteEvaluation(ctx context.Context, req *stock.DeleteEvaluationRequest) (resp *stock.DeleteEvaluationResponse, err error) {
	if mysql.DB == nil {
		return &stock.DeleteEvaluationResponse{Success: false}, nil
	}

	tx := mysql.DB.Begin()
	if tx.Error != nil {
		return &stock.DeleteEvaluationResponse{Success: false}, tx.Error
	}

	var eval model.EvaluationRecord
	if err := tx.Where("id = ?", req.Id).First(&eval).Error; err != nil {
		tx.Rollback()
		return &stock.DeleteEvaluationResponse{Success: false}, nil
	}

	if err := tx.Delete(&eval).Error; err != nil {
		tx.Rollback()
		return &stock.DeleteEvaluationResponse{Success: false}, nil
	}

	if eval.PredictionID != "" {
		if err := tx.Where("id = ?", eval.PredictionID).Delete(&model.PredictionRecord{}).Error; err != nil {
			tx.Rollback()
			return &stock.DeleteEvaluationResponse{Success: false}, nil
		}
	}

	if err := tx.Commit().Error; err != nil {
		return &stock.DeleteEvaluationResponse{Success: false}, err
	}

	return &stock.DeleteEvaluationResponse{Success: true}, nil
}

// GetMarketTrends 实现 StockServiceImpl 接口
func (s *StockServiceImpl) GetMarketTrends(ctx context.Context, req *stock.GetMarketTrendsRequest) (resp *stock.GetMarketTrendsResponse, err error) {
	return &stock.GetMarketTrendsResponse{
		Trends: []*stock.MarketTrend{},
		Total:  0,
	}, nil
}

// UpdateMarketTrend implements StockServiceImpl interface.
func (s *StockServiceImpl) UpdateMarketTrend(ctx context.Context, req *stock.UpdateMarketTrendRequest) (resp *stock.UpdateMarketTrendResponse, err error) {
	return &stock.UpdateMarketTrendResponse{Success: false}, nil
}

// DeleteMarketTrend implements StockServiceImpl interface.
func (s *StockServiceImpl) DeleteMarketTrend(ctx context.Context, req *stock.DeleteMarketTrendRequest) (resp *stock.DeleteMarketTrendResponse, err error) {
	return &stock.DeleteMarketTrendResponse{Success: false}, nil
}

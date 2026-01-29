package provider

import "context"

// 市场数据统一接口
// 约束：
// - code 统一使用 "SH/SZ+6位代码"（示例：SH600000、SZ000001）
// - 所有方法为实时请求，不使用缓存；需要自行控制超时与重试
// - period 支持：1m/5m/15m/30m/60m/day/week/month；返回字段语义一致
type MarketDataClient interface {
	GetIntradayQuote(ctx context.Context, code string) (*IntradayQuote, error)
	GetOrderBook(ctx context.Context, code string, depth int) (*OrderBook, error)
	GetKlines(ctx context.Context, code string, period string, limit int) ([]*KlineItem, error)
}

// 龙虎榜统一接口
// 约束：
// - date 格式为 "YYYY-MM-DD"
// - 返回为统一结构，包含原因与净流入，席位明细通过 GetSeats 获取
type DragonTigerClient interface {
	GetTodayList(ctx context.Context, date string) ([]*DragonTigerItem, error)
	GetHistory(ctx context.Context, code string, limit int) ([]*DragonTigerItem, error)
	GetSeats(ctx context.Context, date string, code string) ([]*SeatRecord, error)
}

// 财务数据统一接口
// 约束：
// - quarters 为返回最近季度数量；字段单位统一为元与比例
type FinancialClient interface {
	GetFinancialReports(ctx context.Context, code string, quarters int) ([]*FinancialReport, error)
}

// 板块统一接口
// 约束：
// - kind 支持 industry/concept；同花顺概念通过 GetStockConceptsTHS 获取
type SectorClient interface {
	GetSectorRank(ctx context.Context, kind string) ([]*SectorRankItem, error)
	GetSectorStocks(ctx context.Context, sectorCode string) ([]*StockBasic, error)
	GetAStockList(ctx context.Context) ([]*StockBasic, error)
	GetStockConceptsTHS(ctx context.Context, code string) ([]*ConceptTag, error)
	GetSectorStocksDetail(ctx context.Context, sectorCode string) ([]*SectorStockDetail, error)
}

// 新闻统一接口
// 约束：
// - 返回时间统一为 UTC 或可解析的时间字符串，需在业务侧做时区处理
type NewsClient interface {
	GetStockNews(ctx context.Context, code string, limit int) ([]*NewsItem, error)
	GetMarketNews(ctx context.Context, limit int) ([]*NewsItem, error)
	GetLiveNews(ctx context.Context, limit int) ([]*LiveNewsItem, error)
}

// 公告统一接口
// 约束：
// - keyword 为可选过滤关键词；返回包含标题、类别、链接与时间
type NoticeClient interface {
	GetStockNotices(ctx context.Context, code string, keyword string, limit int) ([]*NoticeItem, error)
}

// 热度统一接口
// 约束：
// - 返回来源字段指明热度来源（如 东财股吧）
type PopularityClient interface {
	GetStockHeat(ctx context.Context, code string) (*Popularity, error)
}

// 情绪统一接口
// 约束：
// - date 为交易日；数据来源可能为非官方接口，注意稳定性
type SentimentClient interface {
	GetLimitUpPool(ctx context.Context, date string) ([]*LimitUpItem, error)
}

// 行业属性统一接口
// 约束：
// - 返回行业、地域与概念标签文本，来源需标准化
type IndustryClient interface {
	GetIndustryIndex(ctx context.Context, code string) (*IndustryIndex, error)
}

// 筹码分布统一接口
// 约束：
// - 返回平均成本、胜率与 90% 成本区间，单位与比例遵循统一约束
type ChipClient interface {
	GetChipDistribution(ctx context.Context, code string) (*ChipDistribution, error)
}

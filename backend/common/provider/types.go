package provider

import "time"

// 统一类型定义，所有字段遵循中文语义，避免歧义
// 约束：
// - Code 统一使用 "SH/SZ+6位代码"
// - 时间字段统一为 time.Time；若来源为字符串，适配层需转换
// - 金额单位为元，成交量单位为股（如来源为手需折算）

// IntradayQuote 表示盘中实时报价
// 来源：优先东财，失败回退至盘口推导；新浪可作为业务侧备选
type IntradayQuote struct {
	Code         string
	Name         string
	Price        float64
	Open         float64
	High         float64
	Low          float64
	PrevClose    float64
	ChangePct    float64
	Volume       int64
	Amount       float64
	Timestamp    time.Time
}

// OrderBook 五档盘口快照
// 约束：Bids/Asks 固定 5 档；WeiBi/WeiCha 为东财口径
type OrderBook struct {
	Bids [5]Level // 买盘1-5
	Asks [5]Level // 卖盘1-5
	WeiBi  float64
	WeiCha float64
	UpdatedAt time.Time
}

// Level 盘口档位
type Level struct {
	Price  float64
	Volume int64
}

// KlineItem K线数据
// 约束：周期通过接口参数控制；复权策略由适配层处理
type KlineItem struct {
	Time       time.Time
	Open       float64
	Close      float64
	High       float64
	Low        float64
	Volume     int64
	ChangePct  float64
}

// DragonTigerItem 龙虎榜条目
// 约束：Reason 为上榜原因，NetInflow 单位为元
type DragonTigerItem struct {
	Date          string
	Code          string
	Name          string
	ClosePrice    float64
	ChangePercent float64
	Reason        string
	NetInflow     float64
}

// SeatRecord 龙虎榜席位
// 约束：买卖金额单位为元；同一席位可重复出现
type SeatRecord struct {
	SeatName  string
	BuyAmount float64
	SellAmount float64
	NetAmount  float64
}

// FinancialReport 财报摘要
// 约束：YoY 为同比百分比（0-100）
type FinancialReport struct {
	Quarter    string
	Revenue    float64
	NetProfit  float64
	EPS        float64
	YoY        float64
}

// SectorRankItem 板块排行
// 约束：kind 由接口入参决定；NetInflow 单位为元
type SectorRankItem struct {
	Code        string
	Name        string
	ChangePct   float64
	NetInflow   float64
}

// StockBasic 基础股票信息
// 约束：Market/Industry/Concept 为来源标准化后文本
type StockBasic struct {
	Code    string
	Name    string
	Market  string
	Industry string
	Concept  string
}

// SectorStockDetail 板块内股票详情
// 约束：所有金额单位为元；成交量为股
type SectorStockDetail struct {
	Code          string
	Name          string
	Price         float64
	ChangePct     float64
	Volume        int64
	Amount        float64
	MarketCap     float64
}
// NewsItem 新闻
// 约束：PublishedAt 需转换为标准时间；Summary 可能为空
type NewsItem struct {
	Title       string
	Summary     string
	Source      string
	URL         string
	PublishedAt time.Time
}

// LiveNewsItem 快讯
// 约束：富文本由业务层清洗
type LiveNewsItem struct {
	Content     string
	Source      string
	URL         string
	PublishedAt time.Time
}

// NoticeItem 公告
// 约束：Keywords 用于简单过滤；Category 来源分类
type NoticeItem struct {
	Title       string
	Category    string
	URL         string
	PublishedAt time.Time
	Keywords    []string
}

// Popularity 热度数据
// 来源：东财股吧热度；不同来源需标注 Source
type Popularity struct {
	Rank     int
	HeatScore int
	Source   string
}

// LimitUpItem 涨停池条目
// 约束：Boards 为连板数；IsBreak 表示炸板
type LimitUpItem struct {
	Code    string
	Name    string
	Boards  int
	Reason  string
	OpenTime string
	IsBreak  bool
}

// ConceptTag 同花顺概念标签
// 来源：Akshare 同花顺模块；通过本地服务封装
type ConceptTag struct {
	Code   string
	Name   string
	Source string
}

// IndustryIndex 行业属性
// 约束：文本字段为来源标准化值
type IndustryIndex struct {
	IndustryName string
	RegionName   string
	ConceptNames string
}

// ChipDistribution 筹码分布
// 约束：WinnerRate 为百分比（0-100）；成本单位为元
type ChipDistribution struct {
	AverageCost float64
	WinnerRate  float64
	Cost90Low   float64
	Cost90High  float64
}

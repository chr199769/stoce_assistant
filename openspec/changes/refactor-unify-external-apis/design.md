## Context
为降低外部数据源调用的分散与不一致，本设计统一定义面向业务的本地接口层，所有 Provider（eastmoney、sina、sentiment、akshare）置于 common/{provider} 目录，业务仅依赖统一接口，不直接发起第三方 HTTP 请求。

## Goals / Non-Goals
- Goals:
  - 统一盘中数据、龙虎榜、新闻公告、K线、板块与财报的接口定义
  - 通过 Provider 注册表选择实现，横切能力（超时、重试、限流、缓存、结构化中文日志）一致
  - 将 Akshare 能力封装为本地接口供 Go 调用，避免直接依赖 Python 代码
- Non-Goals:
  - 立即移除所有旧调用（采用渐进迁移）
  - 一次性重写 Akshare 全量功能（先通过本地服务封装，再逐步 Go 化）

## 统一接口总览
- MarketDataClient
  - GetIntradayQuote(ctx, code) → IntradayQuote
  - GetOrderBook(ctx, code, depth) → OrderBook
  - GetKlines(ctx, code, period, limit) → []KlineItem
- DragonTigerClient
  - GetTodayList(ctx, date) → []DragonTigerItem
  - GetHistory(ctx, code, limit) → []DragonTigerRecord
  - GetSeats(ctx, date, code) → []SeatRecord
- FinancialClient
  - GetFinancialReports(ctx, code, quarters) → []FinancialReport
- SectorClient
  - GetSectorRank(ctx, kind) → []SectorRankItem
  - GetSectorStocks(ctx, sectorCode) → []StockBasic
  - GetAStockList(ctx) → []StockBasic
  - GetStockConceptsTHS(ctx, code) → []ConceptTag
- NewsClient
  - GetStockNews(ctx, code, limit) → []NewsItem
  - GetMarketNews(ctx, limit) → []NewsItem
  - GetLiveNews(ctx, limit) → []LiveNewsItem
- NoticeClient
  - GetStockNotices(ctx, code, keyword, limit) → []NoticeItem
- PopularityClient
  - GetStockHeat(ctx, code) → Popularity
- SentimentClient
  - GetLimitUpPool(ctx, date) → []LimitUpItem

## 关键类型（规范化）
- Code: 统一为 "市场前缀+代码"（如 "SH600000", "SZ000001"）
- IntradayQuote: { code, time, price, open, high, low, prevClose, change, changePct, volume, amount }
- OrderBook: { bids: [{price, volume}], asks: [{price, volume}], updatedAt }
- KlineItem: { time, open, high, low, close, volume, changePct }
- DragonTigerItem: { code, date, reason, netInflow, turnoverRate }
- DragonTigerRecord: { date, reason, netBuy, netSell }
- SeatRecord: { seatName, buyAmount, sellAmount, netAmount }
- FinancialReport: { quarter, revenue, netProfit, eps, yoy }
- SectorRankItem: { code, name, changePct, netInflow }
- StockBasic: { code, name, market, industry, concept }
- ConceptTag: { code, name, source }
- NewsItem: { title, summary, source, url, publishedAt }
- LiveNewsItem: { content, source, url, publishedAt }
- NoticeItem: { title, category, url, publishedAt, keywords }
- Popularity: { rank, heatScore, source }
- LimitUpItem: { code, name, boards, reason, openTime, isBreak }

## Provider 注册与选择
- ProviderRegistry:
  - Register(name, factory)
  - Resolve(name) → { MarketDataClient?, DragonTigerClient?, ... }
- 默认选择：
  - 行情/盘口/K线 → eastmoney
  - 龙虎榜 → eastmoney
  - 财报 → eastmoney
  - 实时行情补充/新闻 → sina
  - 涨停池 → sentiment
  - 新闻抓取（Akshare）→ akshare（通过本地封装）
- 配置示例：
  - stock_data.provider.market=eastmoney
  - stock_data.provider.news=akshare
  - stock_data.provider.rt=sina

## 数据源选择策略（Akshare vs 东财）
- 原则
  - 实时/高并发/盘口级别数据优先用东财（稳定、QPS 高、字段丰富）
  - 分类与同花顺概念等“知识型”数据由 Akshare 提供（聚合抓取，便于统一）
  - 历史长周期作为补充由 Akshare 兜底，仅在东财失败或数据缺失时使用
  - 始终以统一类型输出；差异通过字段映射和校验消除
- 价格与走势
  - 盘中实时报价 IntradayQuote → 东财（主），新浪（备）；不直接用 Akshare
  - 五档盘口 OrderBook → 东财（主）；Akshare 不参与
  - K 线（日/周/月/5/15/30/60 分）→ 东财（主）；Akshare（备，长历史或东财限流时）
    - 复权策略：统一暴露 adj=none|pre|post；东财参数映射 fqt 处理，Akshare 端按其接口转换
  - 指数 K 线/报价 → 东财（主）
- 龙虎榜
  - 当日与历史龙虎榜、席位明细 → 东财（唯一）
- 财务数据
  - 财报摘要与关键指标 → 东财（主）；Akshare（备，接口不可用时）
- 板块与成分
  - 行业/概念板块排行与成分（东财口径）→ 东财（主）
  - 同花顺概念归属 → Akshare（唯一，通过 /ak/ths/concepts）
- 新闻与快讯
  - 个股新闻 → 东财（主）或 Akshare（备）；统一为 NewsItem
  - 市场新闻与 7x24 → 新浪（主），Akshare（备）
- 热度与情绪
  - 股吧热度 → 东财（主）
  - 涨停池 → Sentiment（东财非官方接口，作为单独 Provider）
- 回退与一致性
  - 回退触发：HTTP 5xx/429、解析失败、字段缺失
  - 一致性校验：每日抽样对比东财与 Akshare 的 K 线收盘价与成交量，阈值 >0.5% 记录告警
  - 统一日志：记录 provider、端点、耗时、重试次数与差异摘要

## 横切策略
- 超时：默认 3s（可配置），K线/财报 5-8s
- 重试：指数退避，最多 3 次，仅对可重试错误（429/5xx/网络抖动）
- 限流：令牌桶/并发上限
- 缓存：不使用缓存，所有数据实时请求
- 日志：结构化中文日志，包含 provider、endpoint、耗时、错误码与关键参数
- 错误码：统一包装为本地错误类型（ProviderError{code, message, endpoint}）

## Akshare 本地封装方案
- 目标：将 Python Akshare 能力通过本地 HTTP 服务暴露，Go 侧作为标准 NewsClient/NoticeClient 的实现，无需直接引入 Python 代码。
- 服务名：akshare_service（本地）
- 端点（初始）：
  - GET /ak/news/stock?symbol=SZ000001&limit=50 → []NewsItem
  - GET /ak/news/market?limit=50 → []NewsItem
  - GET /ak/news/live?limit=50 → []LiveNewsItem
- 端点（THS 概念）：
  - GET /ak/ths/concepts?symbol=SZ000001 → []ConceptTag
  - 说明：内部基于 Akshare 的同花顺模块构建倒排索引
    - 列表：stock_board_concept_name_ths → 概念清单（code/name）
    - 成分：stock_board_concept_cons_ths(concept_code) → 成分股
    - 查询：实时按需计算 symbol→concept 映射，不引入缓存层
- 响应体遵循统一类型；服务内部调用 ak.stock_news_em、ak.news_cctv 等，标准化字段。
- 迁移策略：先接入 HTTP 封装，后续根据需要将功能逐步用 Go 重写或替换数据源。

## 业务调用准则
- 工具层（如 StockAnalysisTool）仅依赖统一接口，不直接创建具体 Provider 客户端。
- 示例（伪代码）：
  - quote := providers.Resolve("market").MarketDataClient.GetIntradayQuote(ctx, "SH600000")
  - lhb := providers.Resolve("lhb").DragonTigerClient.GetHistory(ctx, "SZ000001", 5)
  - news := providers.Resolve("news").NewsClient.GetStockNews(ctx, "SZ000001", 20)

## 迁移映射
- EastMoney：
  - client_extra.go 的盘口、行业属性 → MarketDataClient.GetOrderBook / SectorClient
  - client_kline.go 的 K 线 → MarketDataClient.GetKlines(period=klt)
  - 龙虎榜相关 → DragonTigerClient（List/Seats/History）
  - 财报 → FinancialClient.GetFinancialReports
- Sina：
  - 实时行情 → MarketDataClient.GetIntradayQuote（作为备用/补充）
  - 市场新闻/7x24 → NewsClient.GetMarketNews/GetLiveNews
- Sentiment：
  - 涨停池 → SentimentClient.GetLimitUpPool
- Akshare：
  - 全部通过 akshare_service 接口对齐 NewsClient/NoticeClient 类型

## 风险 / 权衡
- 风险：短期内双轨并存增加复杂度 → 通过 ProviderRegistry 限制调用入口
- 权衡：Akshare 保留 Python 封装 → 快速稳定，但需维护小型服务

## 验证与监控
- 为统一接口添加集成测试（模拟 provider 返回）
- 接入耗时与错误码监控，确保限流与重试效果

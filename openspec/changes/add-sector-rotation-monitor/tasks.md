# 任务列表

## Phase 1: 数据层与基础设施 (Data & Infra)

- [ ] **1.1 数据库设计与实现 (MySQL)**
    - [ ] **设计表结构**:
        - [ ] `market_sector_daily`: 存储每日/每小时的板块热度快照 (字段: `date`, `time`, `sector_code`, `name`, `change_pct`, `net_inflow`, `top_stock`).
        - [ ] `market_limit_up_summary`: 存储每日涨停情绪概览 (字段: `date`, `limit_up_count`, `broken_count`, `highest_board`, `sentiment_score`).
        - [ ] `market_limit_up_detail`: 存储每日涨停个股详情 (字段: `date`, `stock_code`, `stock_name`, `board_count`, `reason`, `is_new`).
    - [ ] **Gorm Model**: 在 `stock_service/dal/model` 中创建对应的 Go Struct。
    - [ ] **Migration**: 编写 SQL 脚本或 AutoMigrate 逻辑。

- [ ] **1.2 缓存策略实现 (Redis)**
    - [ ] **定义 Key 规范**:
        - [ ] `market:sector:rank:{type}` (TTL: 60s) - 实时板块排行。
        - [ ] `market:limit_up:pool` (TTL: 30s) - 实时涨停池数据。
    - [ ] **实现缓存层**:
        - [ ] 在 `stock_service` 中封装 `CacheManager`，处理 `GetOrSet` 逻辑。
        - [ ] 确保高并发下的防击穿处理 (SingleFlight)。

- [ ] **1.3 升级 EastMoney Provider (数据获取)**
    - [ ] 实现 `GetSectorRank`：爬取东方财富板块涨幅榜。
    - [ ] 实现 `GetSectorMoneyFlow`：爬取板块资金流向。
    - [ ] **数据落库**: 实现异步任务或 Cron Job，在收盘后将当日关键数据写入 MySQL (用于历史回测)。

- [ ] **1.4 新增 MarketSentiment Provider (数据获取)**
    - [ ] 实现 `GetLimitUpPool`：爬取涨停、炸板、跌停数据。
    - [ ] **数据清洗**: 解析"连板天数"、"涨停原因"等非结构化文本。
    - [ ] **数据落库**: 收盘后归档当日情绪数据到 MySQL。

- [ ] **1.5 暴露 RPC/HTTP 接口**
    - [ ] IDL 定义 `GetMarketSectors` 和 `GetLimitUpPool`。
    - [ ] Handler 实现：优先读 Redis -> 失败读 Provider -> 异步回写 Redis。

## Phase 2: AI Service 工具与逻辑

- [ ] **2.1 封装 SectorTool**
    - [ ] 在 AI Service 中调用 Stock Service 的新接口。
    - [ ] 为 LLM 提供查询工具：`query_top_sectors` (查询热点), `query_limit_up_summary` (查询情绪)。

- [ ] **2.2 重构新闻工具 (MarketInfoTool)**
    - [ ] 合并 `NewsTool` 和 `TrendTool`。
    - [ ] 增加 `GetGlobalMarkets` (外盘) 支持 (数据源: 新浪财经/雪球 API)。
    - [ ] 统一入口：根据输入(stock_code/trends/global)分发请求。

- [ ] **2.3 开发 MarketReviewAgent**
    - [ ] **Prompt 工程**:
        - [ ] 收盘复盘：整合板块热点、龙头股表现、宏观新闻。
        - [ ] 盘前分析：整合外盘数据、重大消息、昨日情绪。
    - [ ] **Agent 流程**: 实现 `Review` (复盘) 和 `PreMarketAnalysis` (盘前) 的思维链 (CoT)。

- [ ] **2.4 验证与测试**
    - [ ] 单元测试: 测试 Provider 的爬虫解析逻辑。
    - [ ] 集成测试: 验证 API -> Redis -> Provider 的调用链路。
    - [ ] 效果评估: 生成一份样例复盘报告，人工评估准确性和可读性。

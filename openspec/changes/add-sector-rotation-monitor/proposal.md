# 提案：新增板块轮动监控与AI复盘功能

## 1. 背景 (Background)
目前的 `stock_assistant` 系统主要关注个股的基本面和技术面数据，缺乏对**市场整体情绪**和**板块轮动**的监控能力。
参考开源项目 `EasyStock/Bull` (特别是 `src/bankuai` 模块)，发现通过监控"概念板块"、"涨停池"、"连板股"等数据，可以有效捕捉短线市场热点。

## 2. 目标 (Goals)
1.  **板块轮动监控**: 实时获取并分析领涨板块（行业/概念），识别市场热点。
2.  **市场情绪量化**: 通过"涨停池"（首板、连板、炸板）数据，量化当日市场情绪。
3.  **AI 智能复盘与前瞻**: 结合板块数据、涨停数据和新闻资讯，由 AI 生成每日市场复盘报告及盘前分析。

## 3. 变更范围 (Scope)

### Backend
-   **Stock Service**:
    -   扩展 `EastMoney` provider，支持获取板块涨幅榜、资金流向。
    -   新增 `MarketSentiment` provider，获取当日涨停池、连板梯队数据。
    -   新增 API: `GetMarketSectors` (获取板块数据), `GetLimitUpPool` (获取涨停数据)。
-   **AI Service**:
    -   新增 `MarketReviewAgent` (复盘与前瞻助手)。
    -   新增 `SectorTool` (用于 AI 查询板块数据)。
    -   **Refactor**: 合并 `NewsTool` 和 `TrendTool` 为 `MarketInfoTool`，统一新闻资讯入口，并支持查询外盘数据。

### Frontend (Future)
-   新增"市场大盘" Dashboard，展示板块热力图和 AI 复盘摘要。

## 4. 影响 (Impact)
-   **性能**: 爬取板块和涨停数据需要增加一定的网络请求，需配置合理的缓存策略 (Redis)。
-   **数据库**: 需新增表存储历史板块热度和涨停数据，用于回测分析。

## 5. 风险 (Risks)
-   第三方数据接口 (EastMoney) 变动可能导致爬虫失效，需增加容错和重试机制。

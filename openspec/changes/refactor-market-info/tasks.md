## 1. Implementation
- [x] 1.1 **协议与存储**: 在 IDL 中定义 `ProcessMarketTrends` (AI Service) 和 `GetMarketTrends` (Stock Service) 接口，并在 `stock_service` 创建 `MarketTrend` 数据库表。
- [x] 1.2 **AI 服务重构与能力**: 重构 `ai_service` (拆分 `langchain_provider`), 实现 `ProcessMarketTrends` 核心逻辑 (分析/打分/分类)，并改造 `MarketInfoTool` 以调用 RPC。
- [x] 1.3 **Stock 服务 Worker**: 实现完整的 `TrendWorker` (多源抓取 -> RPC调用AI -> 存库 -> 智能清理) 并在 `stock_service` 启动时加载。

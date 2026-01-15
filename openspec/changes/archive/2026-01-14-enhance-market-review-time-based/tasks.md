# Tasks

## 1. AI Service Implementation
- [x] 1.1 在 `MarketReviewAgent` 中实现时间状态检测逻辑 (Pre-market, Intra-day, Post-market)。
- [x] 1.2 实现“盘中分析”的 Prompt 模板和处理逻辑。
- [x] 1.3 扩展 `MarketInfoTool` 或相关工具以支持获取盘中实时数据（如当前涨幅榜、实时资金流）。
- [x] 1.4 更新 AI 服务的 `ReviewMarket` 接口逻辑，支持自动路由。

## 2. Validation
- [x] 2.1 编写单元测试验证时间判定逻辑的正确性。
- [x] 2.2 验证不同时间段请求返回的报告结构是否符合预期。

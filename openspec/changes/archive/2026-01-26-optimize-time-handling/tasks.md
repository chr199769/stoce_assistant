## 1. Implementation
- [x] 1.1 修改 `stock_service` 中的 `eval_worker.go`，实现 `calculateEffectiveDate` 逻辑。
- [x] 1.2 更新 `processEvaluation`，使用 `EffectiveDate` 确定 T+1/T+3 的 K 线索引。
- [x] 1.3 修改 `ai_service` 中的 `AnalyzeMarket`，计算目标预测日期。
- [x] 1.4 更新 `MarketAnalysisMaster` Prompt，支持传入目标日期描述。
- [x] 1.5 验证评估打分逻辑。
- [x] 1.6 验证盘前分析文案生成。

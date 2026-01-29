# Change: 增加预测涨跌幅与增强评分系统

## Why
目前系统仅预测股票的涨跌方向（看涨/看跌）。用户希望了解预测变动的幅度（涨跌幅百分比）。此外，评估评分系统应该奖励准确的涨跌幅预测，而不仅仅是方向。

## What Changes
- 更新 `PredictionRecord` 以存储 `PredictedChange` (float64)。
- 更新 AI Provider 以输出预测的涨跌幅百分比。
- 更新 Evaluation Worker 以基于涨跌幅准确度计算得分。
- **BREAKING**: `prediction_records` 的数据库架构变更。

## Impact
- `stock_service`: 架构更新，评分逻辑更新。
- `ai_service`: Prompt 更新，输出解析更新。

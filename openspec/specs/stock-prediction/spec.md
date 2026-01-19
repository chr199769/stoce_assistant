# stock-prediction Specification

## Purpose
TBD - created by archiving change add-smart-trading-enhancement-suite. Update Purpose after archive.
## Requirements
### Requirement: Fractal Prediction (History Repeats)
系统 MUST 提供基于历史 K 线形态相似度的预测功能。

#### Scenario: Predict future trend
- **WHEN** 用户请求某只股票的走势预测
- **THEN** 系统在历史数据中搜索最相似的 3 个 K 线片段
- **AND** 展示这些历史片段的后续走势作为预测参考
- **AND** 计算相似度得分并展示


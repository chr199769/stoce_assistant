## ADDED Requirements
### Requirement: Percentage Prediction
系统 MUST 预测 T+3 时间范围内的涨跌幅百分比。

#### Scenario: Predict percentage
- **WHEN** 用户请求预测
- **THEN** 系统返回预测的涨跌幅百分比（例如 +5.2% 或 -3.1%）
- **AND** 与该百分比相关的置信度水平

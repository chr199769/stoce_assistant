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

### Requirement: Percentage Prediction
系统 MUST 预测 T+3 时间范围内的涨跌幅百分比。

#### Scenario: Predict percentage
- **WHEN** 用户请求预测
- **THEN** 系统返回预测的涨跌幅百分比（例如 +5.2% 或 -3.1%）
- **AND** 与该百分比相关的置信度水平

### Requirement: 知识图谱增强预测
系统 MUST 在生成股票预测时融合知识图谱证据与多智能体评估结果。

#### Scenario: 融合图谱与协作结论
- **WHEN** 用户请求某只股票的走势预测
- **THEN** 系统查询知识图谱中的公司关系、行业链路与事件影响
- **AND** 汇总多智能体协作的结论与评分
- **AND** 输出预测方向、置信度与证据摘要


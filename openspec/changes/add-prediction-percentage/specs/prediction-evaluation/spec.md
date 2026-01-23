## ADDED Requirements
### Requirement: Accuracy Scoring
系统 MUST 基于方向和涨跌幅准确度来评估预测。

#### Scenario: Calculate score
- **WHEN** T+3 市场数据可用时
- **THEN** 计算预测涨跌幅与实际 ROI 之间的差异
- **AND** 分配一个分数 (0-100)，预测越接近分数越高

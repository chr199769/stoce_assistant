# prediction-evaluation Specification

## Purpose
TBD - created by archiving change add-prediction-percentage. Update Purpose after archive.
## Requirements
### Requirement: Accuracy Scoring
系统 MUST 基于方向和涨跌幅准确度来评估预测。

#### Scenario: Calculate score
- **WHEN** T+3 市场数据可用时
- **THEN** 计算预测涨跌幅与实际 ROI 之间的差异
- **AND** 分配一个分数 (0-100)，预测越接近分数越高

### Requirement: Prediction Accuracy Evaluation
系统 MUST 能够自动评测股票预测的准确性，并根据预测发起的时机正确对齐评估时间窗口。

#### Scenario: Evaluate prediction with correct timing
- **WHEN** AI 生成了一个股票预测
- **THEN** 系统根据生成时间确定**生效日期 (Effective Date)**
  - 若在交易日 15:00 前生成，Effective Date = 当日
  - 若在交易日 15:00 后生成，Effective Date = 下一交易日
- **AND** 系统在 Effective Date + 2 个交易日收盘后进行评估
- **AND** 比较 Effective Date 和 Effective Date + 2 的收盘价变化与预测值


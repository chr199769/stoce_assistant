## ADDED Requirements
### Requirement: Intraday Market Sentinel
系统 MUST 提供实时的市场哨兵服务，分析盘中板块轮动和异常成交量。

#### Scenario: Detect abnormal volume
- **WHEN** 某只股票或板块的成交量显著偏离历史均值（如 > 3倍标准差）
- **THEN** 生成异常成交量警报
- **AND** 包含异常类型（如“底部放量”、“高位出逃”）

#### Scenario: Analyze sector rotation
- **WHEN** 资金流向发生显著变化（从一个板块流向另一个）
- **THEN** 识别并报告板块轮动路径

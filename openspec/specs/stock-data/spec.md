# stock-data Specification

## Purpose
TBD - created by archiving change optimize-stock-prediction. Update Purpose after archive.
## Requirements
### Requirement: 历史龙虎榜
系统 MUST 检索指定股票过去 5 个交易日的龙虎榜数据。

#### Scenario: 最近上榜的股票
- **WHEN** 股票在过去 5 天内出现在龙虎榜上
- **THEN** 系统返回这些上榜的日期、原因和净买入/卖出金额

### Requirement: 技术指标
系统 MUST 检索高级技术指标，包括筹码峰（Chip Peak）分布、当前买卖盘队列和挂单量。

#### Scenario: 获取技术指标
- **WHEN** 请求股票的技术指标
- **THEN** 系统返回当前的筹码分布（获利/亏损比例）、前 5 个买/卖盘和总挂单量

### Requirement: 行业和市场背景
系统 MUST 检索股票的行业分类和当前市场指数信息。

#### Scenario: 获取背景
- **WHEN** 请求股票背景
- **THEN** 系统返回股票的行业板块和相关市场指数（例如上证综指）的当前状态

### Requirement: 公司财报
系统 MUST 检索指定股票的最新财务报告摘要，包括收入、净利润、每股收益 (EPS) 和同比增长率。

#### Scenario: 获取财报摘要
- **WHEN** 用户请求指定股票的财报数据
- **THEN** 系统返回最近 4 个季度的关键财务指标列表


## ADDED Requirements
### Requirement: 公司财报
系统 MUST 检索指定股票的最新财务报告摘要，包括收入、净利润、每股收益 (EPS) 和同比增长率。

#### Scenario: 获取财报摘要
- **WHEN** 用户请求指定股票的财报数据
- **THEN** 系统返回最近 4 个季度的关键财务指标列表

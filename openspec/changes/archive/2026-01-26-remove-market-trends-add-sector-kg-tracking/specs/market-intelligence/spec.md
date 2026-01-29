## REMOVED Requirements
### Requirement: Market Intelligence Gathering
系统 MUST 定期从多个公共来源收集市场趋势和热点信息。

#### Scenario: Scheduled fetching
- **WHEN** 定时任务触发（例如每小时）
- **THEN** 系统从配置的来源（微博、百度等）抓取最新的热搜列表

**Reason**: 市场趋势功能仅用于管理后台，产品侧无强依赖。
**Migration**: 下线对应 Worker、API、页面与存储。

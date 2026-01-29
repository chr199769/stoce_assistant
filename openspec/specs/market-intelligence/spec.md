# market-intelligence Specification

## Purpose
TBD - created by archiving change refactor-market-info. Update Purpose after archive.
## Requirements
### Requirement: Intelligent Filtering
系统 MUST 使用 AI 模型分析收集到的信息，过滤掉非金融相关的内容。

#### Scenario: Filter entertainment news
- **WHEN** 抓取到一条关于"某明星绯闻"的热搜
- **THEN** AI 分析判断其金融相关性为低，该条目被丢弃或标记为无关

#### Scenario: Identify market impact
- **WHEN** 抓取到一条关于"半导体行业补贴政策"的热搜
- **THEN** AI 识别其为高金融相关性，提取关键摘要，并关联"半导体"板块

### Requirement: Impact Classification
系统 MUST 将市场情报分类为“长期政策支持”或“短期消息影响”，并赋予不同的权重。

#### Scenario: Long-term Policy
- **WHEN** 抓取到“央行宣布降准”的消息
- **THEN** 系统将其分类为 `policy_long_term`，并赋予较高权重（如 2.0），有效期设置为较长（如 30 天）。

#### Scenario: Short-term News
- **WHEN** 抓取到“某公司中标合同”的消息
- **THEN** 系统将其分类为 `short_term_news`，并赋予标准权重（如 1.0），有效期设置为较短（如 24 小时）。

### Requirement: Trend Storage and Retrieval
系统 MUST 在 Stock Service 中存储经过筛选的市场情报，并支持通过 RPC 检索。系统 MUST 通过 AI 复核自动清理过时信息。

#### Scenario: Tool query
- **WHEN** 预测 Agent 调用市场情报工具
- **THEN** 工具通过 RPC 请求 Stock Service，返回有效的长期政策和最近的短期消息。

#### Scenario: Smart cleanup
- **WHEN** 消息超过初筛时间（如 24h）
- **THEN** 系统调用 AI 评估其当前影响力。
- **IF** AI 判定影响已消退，**THEN** 删除或归档该消息。
- **IF** AI 判定仍有影响，**THEN** 保留该消息。


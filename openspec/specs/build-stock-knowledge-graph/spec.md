# build-stock-knowledge-graph Specification

## Purpose
TBD - created by archiving change add-stock-prediction-kg-multi-agent. Update Purpose after archive.
## Requirements
### Requirement: 股票知识图谱构建
系统 MUST 从公开数据源构建股票知识图谱，并保留实体与关系的来源信息。

#### Scenario: 构建图谱主干
- **WHEN** 系统接收到日终数据更新
- **THEN** 生成公司、行业、概念、事件等实体与关系
- **AND** 为每条关系记录来源与更新时间

### Requirement: 知识图谱查询
系统 MUST 提供按股票、行业与事件的图谱查询能力。

#### Scenario: 按股票查询关系
- **WHEN** 用户或智能体请求某只股票的关系网络
- **THEN** 返回相关的上游供应链、同行业公司与近期事件
- **AND** 附带关系强度与证据来源

### Requirement: 图谱运维可视化
系统 MUST 提供知识图谱运维页面用于查询实体、关系与事件详情。

#### Scenario: 运维查看图谱数据
- **WHEN** 运维人员在管理端查询某实体或关系
- **THEN** 系统展示关联实体、关系强度与证据来源
- **AND** 支持按时间范围与可信度筛选

### Requirement: 热门板块核心股票跟踪
系统 MUST 维护指定热门板块的核心股票清单，并将其写入知识图谱用于跟踪。

#### Scenario: 获取热门板块核心股票
- **WHEN** 系统刷新板块清单
- **THEN** 使用公开板块数据匹配以下板块名称并选择核心股票：商业航天、可控核聚变、光伏、液冷、CPO、AI、能源、算力、芯片、存储、机器人、光刻机、证券、银行、竣工
- **AND** 每个板块至少选取 10 只股票

#### Scenario: 写入知识图谱
- **WHEN** 核心股票清单生成
- **THEN** 系统将板块与股票关系写入图谱并记录更新时间


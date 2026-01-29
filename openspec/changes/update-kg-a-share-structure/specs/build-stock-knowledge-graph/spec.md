## ADDED Requirements
### Requirement: A股图谱主干重建
系统 MUST 支持清空现有知识图谱并基于全量数据重建主干结构。

#### Scenario: 执行重建流程
- **WHEN** 运维执行图谱重建
- **THEN** 系统清空现有实体与关系
- **AND** 重新写入市场、板块与股票主干结构

### Requirement: 市场与交易板块实体
系统 MUST 写入A股市场实体与交易板块实体，并建立市场关系。

#### Scenario: 导入交易板块
- **WHEN** 系统开始重建
- **THEN** 写入市场实体“ A股市场 ”
- **AND** 写入交易板块实体（上证、深证、创业板、科创、北证）
- **AND** 建立市场与交易板块关系

### Requirement: 行业板块与概念板块实体
系统 MUST 导入全量行业板块与概念板块，并可区分板块类型。

#### Scenario: 导入全量板块
- **WHEN** 系统拉取板块列表
- **THEN** 写入行业板块与概念板块实体
- **AND** 使用属性标记板块类型
- **AND** 建立板块与市场关系

### Requirement: 板块成分股与关系标注
系统 MUST 导入板块成分股并标注核心股、龙头股与带动信息。

#### Scenario: 导入成分股关系
- **WHEN** 系统拉取板块成分股
- **THEN** 过滤包含“ST”的股票
- **AND** 写入股票实体
- **AND** 建立板块与股票关系
- **AND** 在关系证据中记录核心股、龙头股、带动信息与更新时间

### Requirement: 股票交易板块归属
系统 MUST 建立股票与交易板块的归属关系。

#### Scenario: 关联股票所属交易板块
- **WHEN** 股票实体被写入
- **THEN** 建立股票与交易板块关系

### Requirement: 定时采集（分析模式）
系统 MUST 定时获取市场消息与时政新闻，并分析其对板块与股票的影响后写入图谱。

#### Scenario: 采集与分析
- **WHEN** 定时任务触发（建议每小时）
- **THEN** 从公开来源抓取市场/时政新闻
- **AND** 调用 AI 服务进行归因与影响分析（板块/股票）
- **AND** 以 Event(REPORT/NEWS/POLICY) + Relation(AFFECTS/BENEFITS) + Evidence 写入图谱
- **AND** 为每条事件生成去重键、可信度与更新时间

### Requirement: 定时采集（快照模式）
系统 MUST 定时获取指定股票/板块的“财务、估值、公告”等结构化信息，并直接写入图谱。

#### Scenario: 快照写入
- **WHEN** 定时任务触发（财务/估值建议每日；公告建议每4小时）
- **THEN** 拉取财务快照（营收/净利/ROE/现金流等）、估值快照（PE/PB/PS/分位等）、公告（预告/快报/重大事项等）
- **AND** 使用 Event(REPORT) + Relation(BELONGS_TO/AFFECTS) + Evidence(category=financial/valuation/notice) 表示
- **AND** 直接写入实体关联（不进行智能分析），保留来源、时间戳与置信度
- **AND** 支持覆盖/去重策略，保证幂等

## MODIFIED Requirements
### Requirement: 热门板块核心股票跟踪
系统 MUST 维护全量板块成分股，并在板块-股票关系中标注核心股、龙头股与带动信息。

#### Scenario: 全量板块关系写入
- **WHEN** 系统完成板块成分股导入
- **THEN** 每个板块的股票关系包含核心股、龙头股、带动标记

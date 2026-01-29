# 设计：A股知识图谱重建与定时采集双模式

## 总览
系统采用“重建主干 + 定时采集双模式”以保证图谱的完整性与时效性：
- 重建主干：一次性写入市场/交易板块/行业板块/概念板块/成分股与归属关系
- 定时采集：
  - 分析模式：新闻/时政 → AI 归因 → 事件/关系 + 证据
  - 快照模式：财务/估值/公告 → 直接写入 Event(REPORT) + Evidence + 关系

## 实体与关系映射
- 市场实体：`Entity(id="market:a-share", type=SECTOR, name="A股市场")`
- 交易板块：`Entity(type=SECTOR)` 上证/深证/创业板/科创/北证
- 行业/概念板块：`Entity(type=SECTOR, attributes.type="industry|concept")`
- 股票：`Entity(type=STOCK, attributes.code=实际代码)`
- 关系：
  - 市场 → 交易板块：`Relation(type=BELONGS_TO)`
  - 市场 → 行业/概念板块：`Relation(type=BELONGS_TO)`
  - 股票 → 交易板块：`Relation(type=BELONGS_TO)`
  - 股票 ↔ 板块：`Relation(type=BELONGS_TO)`，证据中记录核心股/龙头股/带动
  - 事件 → 股票/板块：`Relation(type=AFFECTS|BENEFITS)`，来源于分析或快照

## 分析模式（TrendWorker）
- 周期：每小时
- 数据源：公开新闻/政策/市场趋势（如 EastMoney、Sina），统一聚合
- 过程：
  1. 抓取原始文本与元信息（时间、URL、来源）
  2. 调用 AI 服务 `ProcessMarketTrends` 做归因：关联板块/股票、影响方向与摘要
  3. 写入 `Event(NEWS|POLICY|REPORT)`，生成 `dedupe_key`（hash(title|source|timestamp)）
  4. 写入 `Relation(AFFECTS|BENEFITS)` 指向板块/股票，附带 `Evidence(category="trend")`
- 幂等：
  - 事件 `Event.id=dedupe_key`，重复写入覆盖更新时间
  - 关系以 `(source_id,target_id,type)` 唯一，更新 `updated_at`

## 快照模式（SnapshotWorker）
- 周期：财务/估值每日，公告每4小时
- 数据源：公开接口（EastMoney等），结构化字段
- 过程：
  1. 拉取财务/估值/公告，转为 `Event(type=REPORT)`，分别用 `Evidence.category=financial|valuation|notice`
  2. 对于公告，作为 `Event(REPORT)` 关联到股票/板块；对于财务/估值，使用 `Relation(BELONGS_TO)` 将快照事件挂到股票
  3. 记录 `timestamp`、`confidence`（数据完整性评分）、`source`
- 幂等：
  - 以 `(symbol, category, period)` 生成 `dedupe_key`，保证覆盖更新

## 数据质量与过滤
- ST 过滤：名称包含“ST”或“退”，以及新股（名称前缀“N”“C”）在关系侧过滤
- 分页覆盖：板块成分股拉取需分页以覆盖 >100 的场景
- 证据完整性评分：缺关键字段时降级并打点

## 运行与监控
- 日志：中文日志，包含数据源、实体/关系写入数量、失败原因
- 指标：写入成功率、事件去重命中率、每周期新增关系数
- 回滚：重建前做快照（导出节点与关系计数），异常时回退到前一版本

## 安全与性能
- 速率限制：对外部接口做限速与重试
- 并发：分页并发拉取 + 图谱写入批处理
- 隐私：仅使用公开市场数据

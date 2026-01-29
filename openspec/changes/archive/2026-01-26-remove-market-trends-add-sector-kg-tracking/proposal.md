# Change: 移除市场趋势功能并新增板块核心股票图谱跟踪

## Why
市场趋势功能主要被管理后台使用，前台与智能体链路依赖较弱。新增对热门板块与核心股票的图谱跟踪，有助于提升后续分析与事件关联质量。

## What Changes
- **BREAKING** 移除市场趋势采集、AI 处理、存储与管理后台页面
- 新增指定热门板块的核心股票清单获取与图谱跟踪

## Impact
- 受影响的规范: market-intelligence, build-stock-knowledge-graph
- 受影响的代码: Stock Service 趋势 Worker 与接口、AI Service 趋势处理、Gateway 市场趋势接口、管理后台市场趋势页面、知识图谱入库 Worker

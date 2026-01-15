# Implementation Plan

## Phase 1: Research & Core Logic (Non-destructive)
- [ ] 1.1 校验现有龙虎榜接口可用性，确保能获取席位详情（API Validation）。
- [ ] 1.2 升级 `SectorTool`：
    - 实现 `GetSectorStocksRaw`：获取板块成份股原始数据。
    - 实现龙头筛选算法（Leader Selection Logic）。
- [ ] 1.3 增强 `DragonTigerTool`：
    - 实现 `GetDragonTigerList`：获取每日榜单。
    - 实现 `MapSeatToHotMoney`：席位到游资的映射逻辑。

## Phase 2: Interface & Service (Contract Changes)
- [ ] 2.1 更新 IDL 定义：
    - `idl/stock.thrift`: 定义 `SectorDetail`, `DragonTigerItem` 等结构及 Service 方法。
    - `idl/api.thrift`: 定义对应的 API 结构及路由。
- [ ] 2.2 实现后端 Service 逻辑 (`StockService`)：
    - 实现 `GetSectorDetail` 方法。
    - 实现 `GetDragonTigerList` 方法。
- [ ] 2.3 实现 Gateway 路由与 Handler：
    - `/api/sector/:sector_code/detail`
    - `/api/dragon-tiger/list`

## Phase 3: AI Logic Enhancement (Business Logic)
- [ ] 3.1 优化 `PredictionProvider`：
    - 更新 Prompt，加入“板块共振”因子。
    - 集成 `SectorTool` 数据到预测上下文中。

## Phase 4: Frontend Implementation (Visible Changes)
- [ ] 4.1 开发板块详情页 (`SectorDetailScreen.tsx`)：
    - 展示龙头股（带标签）、成份股列表。
- [ ] 4.2 开发龙虎榜页 (`DragonTigerScreen.tsx`)：
    - 展示每日上榜个股及席位详情。
- [ ] 4.3 集成与导航 (`AppNavigator`, `SummaryScreen`)：
    - 添加导航路由。
    - 在大盘总结页添加入口。
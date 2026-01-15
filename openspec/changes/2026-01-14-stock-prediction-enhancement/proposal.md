# Change: Stock Prediction Enhancement

## Why
用户需要更精准的选股工具和更深入的市场数据分析能力。
当前的预测模型缺乏板块与个股的共振分析，导致预测准确度受限。
同时，缺乏板块详情页和每日龙虎榜页面，限制了用户对市场热点和资金流向的深入洞察。

## What Changes
- **增强预测逻辑**：在 `PredictionProvider` 中引入板块-个股共振因子，提高选股准确性。
- **新增板块详情 API**：支持获取板块排名、成份股列表以及识别板块龙头股。
- **新增龙虎榜 API**：支持获取每日龙虎榜榜单，并增加席位与游资（HotMoney）的映射分析。
- **前端页面更新**：新增板块详情页和龙虎榜页面，展示上述数据。

## Impact
- **stock-data**:
  - 修改 `Sector Rotation Data` 需求，增加板块详情和成份股支持。
  - 修改 `历史龙虎榜` 需求，扩展为 `Dragon Tiger List Data`，支持每日榜单和席位分析。
  - 新增 `Stock Prediction Resonance` 需求，定义共振预测能力。
- **UI/UX**: 新增两个页面，优化预测展示。

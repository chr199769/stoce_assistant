# Change: Refactor Home and Analysis Features

## Why
当前首页功能分散，个股预测与自选股列表分离，用户体验不佳。
同时，盘前分析和复盘总结混合在其他功能中，缺乏独立入口，且分析深度不足，缺乏热门股票和推荐信息。

## What Changes
- **Mobile**:
  - 创建独立的“市场分析”页面，包含盘前分析和复盘总结。
  - 重构首页，将“个股预测”功能整合到自选股列表中。
- **Backend (AI Service)**:
  - 优化分析 Prompt，增加热门股票和推荐股票的获取逻辑。

## Impact
- 受影响的规范: `analyze-market`, `track-stocks`
- 受影响的代码: 移动端首页组件、分析页面组件、AI 服务 Prompt 模板。

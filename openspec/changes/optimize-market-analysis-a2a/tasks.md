# Tasks

- [ ] 修改 `MarketAnalysisMaster` 提示词 (defaults.go) @ai
  - 增加推荐数量至 6 只
  - 明确初筛定位
- [ ] 实现 A2A 编排逻辑 (handler.go) @ai
  - 解析 Analyst 推荐列表
  - 并发调用 Predictor Agent
  - 实现过滤和排序算法
- [ ] 验证 A2A 流程 @ai
  - 运行服务
  - 模拟请求
  - 检查日志确认 A2A 调用链

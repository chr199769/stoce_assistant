## 1. Implementation
- [ ] 1.1 Backend: 在 `ai_service` 中集成 Langfuse SDK，替换/增强现有的 Prompt 管理。
- [ ] 1.2 Backend: 在 `stock_service` 中设计数据库表 `prediction_evaluations` 用于存储预测快照、实际股价和评分。
- [ ] 1.3 Backend: 在 `stock_service` 实现定时任务，每日盘后更新 T+1, T+2, T+3 的股价并计算准确率得分。
- [ ] 1.4 Backend: 实现评测 API (List, Detail, Statistics)。
- [ ] 1.5 Frontend: 初始化 `web_admin` 项目 (React + Ant Design)。
- [ ] 1.6 Frontend: 实现评测列表页和详情页，展示 Prompt、预测结果、实际走势对比图。
- [ ] 1.7 Ops: 配置 Langfuse 环境变量和项目设置。

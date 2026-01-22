# Change: Add Prediction Evaluation Mechanism

## Why
目前系统缺乏对股票预测结果的自动评测机制，无法量化 AI 预测的准确性，也难以进行针对性的 Prompt 优化。用户无法直观看到历史预测效果，影响对系统的信任度。

## What Changes
- 引入 Langfuse 进行 Prompt 管理和预测链路追踪 (Tracing)。
- 新增“预测评测”能力，自动跟踪预测后 3 天的股价走势，并与预测结果对比打分。
- 新增独立的 Admin 前端页面，用于展示评测记录、分数和偏差分析。
- 支持基于评测结果的 Prompt 动态调整建议。

## Impact
- 受影响的规范: `prediction-evaluation` (新增)
- 受影响的代码:
    - `backend/ai_service`: 集成 Langfuse，记录预测 Trace。
    - `backend/stock_service`: 新增评测数据存储和定时任务 (Cron Job) 用于更新股价和计算分数。
    - `frontend/admin`: 新增 Web 端管理后台 (React)。

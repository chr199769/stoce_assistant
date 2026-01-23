# Change: Optimize Time Error Handling

## Why
目前系统在处理“评估打分”和“盘前分析”时存在时间误差问题：
1. 评估打分：未区分预测发起时间（盘前/盘中 vs 盘后），导致基准日期（T日）计算错误。
2. 盘前分析：生成的文案未根据当前时间动态调整（如“明天” vs “今天”），导致用户困惑。

## What Changes
- 引入“生效日期 (Effective Date)”概念：
  - 盘前/盘中 (收盘前) 发起的预测，生效日 = 当日。
  - 盘后发起的预测，生效日 = 下一交易日。
- 修改评估打分逻辑：统一基于 Effective Date 计算打分区间。
- 修改盘前分析 Prompt：注入目标日期描述。

## Impact
- 受影响的规范: `prediction-evaluation`, `ai-review`
- 受影响的代码: `stock_service` (eval_worker), `ai_service` (analyst)

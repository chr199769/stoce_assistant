## Context
用户希望建立闭环的优化机制：预测 -> 跟踪 -> 评测 -> 优化 Prompt。需要引入外部工具 Langfuse 来专业化管理 LLM 相关的 Trace 和 Prompt。

## Goals / Non-Goals
- Goals:
    - 自动化评分：无需人工干预，系统根据收盘价自动计算偏差。
    - Prompt 版本化：通过 Langfuse 管理 Prompt 版本，方便回滚和 A/B 测试。
    - 可视化：提供独立的后台页面查看效果。
- Non-Goals:
    - 自动修改 Prompt：系统只提供分析和建议，不直接自动修改线上 Prompt（需要人工确认）。

## Decisions
- Decision: 使用 Langfuse 作为 Trace 和 Prompt 管理平台。
    - Why: 开源、功能完善、支持 Go SDK。
- Decision: 评测逻辑放在 `stock_service` 的定时任务中。
    - Why: `stock_service` 拥有股价数据，方便进行对比计算。`ai_service` 只负责生成预测。
- Decision: 新建 `web_admin` 作为独立的前端项目。
    - Why: 评测数据主要面向开发者和运营人员，不适合放在面向 C 端用户的 App 中。

## Risks / Trade-offs
- Risk: Langfuse 服务稳定性。
    - Mitigation: 预测主流程应降级处理，如果 Langfuse 不可用，降级为本地默认 Prompt，不影响用户使用。

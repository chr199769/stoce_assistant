## ADDED Requirements
### Requirement: Prediction Evaluation
系统 MUST 能够自动评测股票预测的准确性。

#### Scenario: Track prediction accuracy
- **WHEN** AI 生成了一个股票预测
- **THEN** 系统记录预测时的目标价、止损价和日期
- **AND** 在随后的 3 个交易日收盘后，系统自动记录实际收盘价
- **AND** 计算预测偏差率和得分

### Requirement: Langfuse Integration
系统 MUST 使用 Langfuse 管理 Prompt 和记录 Trace。

#### Scenario: Prompt management
- **WHEN** 管理员在 Langfuse 中更新了 Prompt
- **THEN** `ai_service` 在下一次请求时应使用最新的 Prompt (或配置的特定版本)

### Requirement: Admin Dashboard
系统 MUST 提供一个 Web 界面用于查看评测结果。

#### Scenario: View evaluation history
- **WHEN** 管理员访问评测页面
- **THEN** 显示历史预测列表，包含预测日期、股票代码、预测结论、实际走势和得分

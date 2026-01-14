< ACTION >
当用户输入以下指令时，你需要按照执行指令对应的操作
1. /do 读取文件内容，按照模板步骤和项目上下文执行，该命令一定是用户选择了openspec中的某个变更
2. /feat 使用openspec流程创建一个变更提案，严格遵循OpenSpec Instructions，用户会明确指定变更内容，例如：/feat 新增股票选股功能
3. /arch 执行openspec的归档功能，将所有变更都归档

<!-- OPENSPEC:START -->
# OpenSpec 指南

这些指南适用于在该项目中工作的 AI 助手。

当请求满足以下条件时，请务必打开 `@/openspec/AGENTS.md`：
- 提及规划或提案（如 proposal, spec, change, plan 等词汇）
- 引入新功能、破坏性变更、架构调整或重大的性能/安全工作
- 听起来模棱两可，你在编码前需要权威的规范

使用 `@/openspec/AGENTS.md` 来了解：
- 如何创建和应用变更提案
- 规范格式和约定
- 项目结构和准则

保留此管理块，以便 'openspec update' 可以刷新指南。

<!-- OPENSPEC:END -->


< MORE IMPORTANT>
1. 使用中文答复
2. 严禁使用内部依赖以及各种非开源依赖！
3. 在编写变更、总结、任务时，除了openspec会校验的关键词（如`proposal`, `change`, `spec`, `MUST`, `SHOULD`等），都必须使用中文。你需要严格保相关文件可以通过openspec校验。

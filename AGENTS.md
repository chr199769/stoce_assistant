<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

< MORE IMPORTANT>
1. 使用中文答复
2. 严禁使用内部依赖以及各种非开源依赖！
3. 日志、注释必须使用中文，保持注释简洁
4. 每次修改之后，要严格确保修改的服务可以正常运行并且正在运行，检查服务依赖是否包含内部依赖
5. 在编写变更、总结、任务时，除了openspec会校验的关键词（如`proposal`, `change`, `spec`, `MUST`, `SHOULD`等），都必须使用中文。你需要严格保相关文件可以通过openspec校验。
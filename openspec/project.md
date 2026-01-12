# 项目背景

## 目标
股票助手（Stock Assistant）是一款专为个人 A 股投资者设计的智能监控和分析工具。它通过实时监控、智能分析和 AI 预测，帮助用户实时掌握市场动态，捕捉投资机会。作为一款个人工具，它无需注册或登录即可使用。

## 技术栈
- **移动端**: React Native (0.83.1), TypeScript (^5.8.3), React (19.2.0)
  - **UI 组件库**: react-native-paper (^5.14.5), react-native-chart-kit (^6.12.0), react-native-vector-icons (^10.3.0)
  - **导航**: React Navigation v7 (@react-navigation/native, bottom-tabs, native-stack)
  - **本地存储**: @react-native-async-storage/async-storage (^2.2.0)
  - **网络请求**: Axios (^1.13.2)
- **后端框架**: 
  - **HTTP 网关**: CloudWeGo Hertz (v0.10.3)
  - **RPC 微服务**: CloudWeGo Kitex (v0.15.4)
- **编程语言**: Go (1.24.11), TypeScript
- **AI 框架**: LangChainGo (v0.1.14)
- **数据库**: PostgreSQL (计划中), Redis (计划中) - 当前主要依赖实时接口和内存
- **IDL**: Thrift

## 项目规范

### 代码风格
- **前端**:
  - 移动端代码全部使用 TypeScript。
  - 使用带 Hooks 的函数式组件。
  - 使用 Prettier (v2.8.8) 进行格式化，ESLint 进行代码检查。
  - UI 文本必须使用中文。
- **后端**:
  - 遵循标准 Go 语言习惯和 CloudWeGo 最佳实践。
  - 使用 `kitex` / `hz` 工具进行代码生成。
  - 禁止使用字节跳动内部依赖（必须使用公开的 GitHub/npm 包）。

### 架构模式
- **微服务架构**:
  - **网关 (Gateway)**: 基于 Hertz 的 HTTP 网关，向移动端暴露 REST API。
  - **AI 服务 (AI Service)**: 基于 Kitex 的 RPC 服务，集成多种 LLM 模型（OpenAI, Zhipu, Qwen, Doubao, DeepSeek）。
  - **股票服务 (Stock Service)**: 基于 Kitex 的 RPC 服务，对接新浪财经等外部数据源。
- **通信**: 客户端与网关之间使用 HTTP/JSON；网关与服务之间使用 Thrift/Kitex。
- **数据流**: 移动端 -> 网关 -> 服务 -> (外部 API/LLM)。

### 测试策略
- **前端**: 使用 Jest 进行单元测试和组件测试。
- **后端**: 使用 Go 标准库 `testing` 包，配合 `testify` 进行断言。
- **验证**: 涉及 API 的业务逻辑变更必须通过本地测试进行验证。

### Git 工作流
- 功能分支工作流。
- 提交信息应原子化且描述清晰。

## 领域背景
- **市场**: 中国 A 股（上海/深圳）。
- **关键实体**: 股票（代码、名称、价格）、板块、大盘指数。
- **功能**: 实时价格监控、每日市场总结、基于 AI 的价格预测。

## 重要约束
- **无内部依赖**: 严格使用公开的开源库。
- **隐私**: 无需用户注册；优先使用本地存储保存用户偏好。
- **语言**: 应用内容必须为中文。
- **性能**: 移动端优先设计，针对 Android 设备进行优化。

## 外部依赖
- **市场数据**: 新浪财经 API (Sina Finance)。
- **AI 模型**: 
  - OpenAI (GPT-4o)
  - Zhipu AI (GLM-4.6v-flash)
  - Aliyun Qwen (Qwen-Turbo)
  - Volcengine Doubao (Doubao-Pro-4k)
  - DeepSeek (DeepSeek-Chat)

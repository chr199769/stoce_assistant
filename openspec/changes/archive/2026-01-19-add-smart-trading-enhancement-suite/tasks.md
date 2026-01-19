## 1. Database & IDL (Foundation)
- [x] 1.1 `stock_service`: 定义 `UserWatchlist` GORM 模型 (`dal/model/watchlist.go`)
- [x] 1.2 `stock_service`: 定义 `IntradaySignal` GORM 模型 (`dal/model/signal.go`)
- [x] 1.3 `stock_service`: **[New]** 定义 `User` GORM 模型 (`dal/model/user.go`)
    - Fields: `ID`, `Username` (unique), `CreatedAt`
- [x] 1.4 `stock_service`: 更新 `dal/mysql/init.go` 中的 `AutoMigrate` 列表
- [x] 1.5 `idl`: 更新 `stock.thrift` 定义 RPC 接口
    - `GetOrCreateUser(username)`
    - `AddWatchlist`, `SaveIntradaySignal`, `GetHistoricalKline`
- [x] 1.6 `script`: 运行 `kitex` 命令重新生成代码
- [x] 1.7 `infrastructure`: 编写 `docker-compose.yml` 添加 `neo4j` 和 `langfuse` 服务配置

## 2. Stock Service Implementation
- [x] 2.1 **[New]** 实现 `User` 相关的 Handler (`GetOrCreateUser`)
- [x] 2.2 实现 `Watchlist` 相关的 CRUD Handler (需支持 `user_id` 参数)
- [x] 2.3 实现 `IntradaySignal` 相关的读写 Handler
- [x] 2.4 优化 `GetHistoricalKline` 接口
- [x] 2.5 实现 `EvalWorker`: 定时任务，从 Langfuse 获取历史预测 -> 查询实际行情 -> 回填 Accuracy Score

## 3. AI Service Implementation (Core Intelligence)
- [x] 3.1 实现 `IntradaySentinel` Agent (异常量/板块轮动)
- [x] 3.2 实现 `FractalPatternMatcher` (K线相似度)
- [x] 3.3 集成 `Neo4jProvider` (Skipped/Reverted due to complexity):
    - 实现 `GraphRetriever` 工具
    - 初始化 Neo4j 连接与 Schema
- [x] 3.4 集成 `LangfuseProvider`:
    - 封装 `TraceManager`
    - 记录 Input, Output, Metadata

## 4. Gateway & Mobile Implementation
- [x] 4.1 `gateway`: 暴露 `/api/user/login` (调用 `GetOrCreateUser`)
- [x] 4.2 `gateway`: 暴露 `/api/watchlist` 等接口 (需从 Header 获取 `X-User-ID`)
- [x] 4.3 `mobile`: **[New]** 新增 `LoginScreen` (输入用户名 -> 存本地 -> 导航至主页)
- [x] 4.4 `mobile`: 改造 `WatchlistScreen` 对接云端数据
- [x] 4.5 `mobile`: 新增 `FractalChart` 组件

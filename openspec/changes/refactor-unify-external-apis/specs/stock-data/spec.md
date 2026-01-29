## ADDED Requirements
### Requirement: 统一外部数据源封装与调用
系统 MUST 将所有第三方数据源客户端集中在 common/{provider}/ 目录，并为每个外部依赖提供覆盖常用接口的统一 Client 封装，包含超时、重试、限流、缓存与结构化中文日志。

#### Scenario: 新增数据源封装
- **WHEN** 引入新的外部数据源
- **THEN** 在 common/{provider}/ 下创建对应的 Client 封装，覆盖我们需要的接口
- **AND** 封装统一包含超时、重试、限流、缓存与结构化中文日志

#### Scenario: 业务层调用统一封装
- **WHEN** 业务模块需要获取股票 K 线、财报或技术指标
- **THEN** 通过统一的 Provider Client 进行调用
- **AND** 不直接在业务层编写第三方 HTTP 请求逻辑

#### Scenario: 失败重试与错误返回
- **WHEN** 外部数据源返回可重试错误或超时
- **THEN** 统一封装按指数退避策略进行有限次重试
- **AND** 超过重试上限后返回可解释错误并记录结构化中文日志

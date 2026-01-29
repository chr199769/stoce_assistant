# Design: Market Intelligence Refactor

## Context
目前的 `MarketInfo` 工具在用户请求时实时抓取外部网站（微博、百度等）的热搜。这种方式存在高延迟、不稳定性（反爬虫、接口变动）以及数据噪声大（包含大量娱乐新闻）的问题。这降低了 AI 股票分析的准确性和用户体验。

## Goals
- **降低延迟**：将工具响应时间从几秒/几十秒降低到毫秒级。
- **提高质量**：通过 LLM 预处理，只保留金融相关的热点，并提取关键信息。
- **增强稳定性**：后台异步抓取，失败重试，不影响用户主流程。
- **结构化数据**：将非结构化的热搜文本转换为结构化的金融情报（包含相关板块、情感倾向等）。

## Decisions

### 1. 架构分离：异步 Worker + 数据库
- **Decision**: 移除 `MarketInfo` 工具中的实时抓取逻辑。引入 `TrendWorker` 在后台定期运行。
- **Rationale**: 实时抓取不可靠且慢。后台抓取可以控制频率，进行重试，并进行耗时的 LLM 处理。
- **Data Flow**:
  1. `TrendWorker` (Cron Job in **Stock Service**) -> Fetch Raw Data (Weibo/Baidu/etc.)
  2. `Stock Service` -> RPC `ai_service.ProcessMarketTrends` -> Structured Market Trends
  3. `Stock Service` -> Save to Database (`market_trends` table in `stock_service`)
  4. User Request -> `MarketInfoTool` (in `ai_service`) -> RPC `stock_service.GetMarketTrends`

### 2. 数据库设计
- **Decision**: 在 `stock_service` 中存储市场情报。
- **Schema**:
  ```sql
  CREATE TABLE market_trends (
      id BIGINT PRIMARY KEY AUTO_INCREMENT,
      source VARCHAR(32) NOT NULL,
      title VARCHAR(255) NOT NULL,
      summary TEXT,
      original_url VARCHAR(512),
      financial_relevance INT,
      related_sectors JSON,
      sentiment_score FLOAT,
      impact_type VARCHAR(20) DEFAULT 'short_term_news',
      weight FLOAT DEFAULT 1.0,
      is_still_valid BOOLEAN DEFAULT TRUE, -- AI 复核标志
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
      INDEX idx_created_relevance (created_at, financial_relevance),
      INDEX idx_validity (is_still_valid)
  );
  ```

### 3. LLM 处理流程 (AI Service)
- **Decision**: `ai_service` 暴露无状态 RPC 接口 `ProcessMarketTrends`。
- **Logic**: 接收原始文本列表，返回结构化分析结果。不直接操作数据库。

### 4. 智能清理策略 (Smart Cleanup)
- **Decision**: 废弃单纯基于时间的清理。引入 AI 相关性复核。
- **Policy**:
  - **初筛**: 超过一定时间（如 24h）的消息标记为 "待复核"。
  - **AI 复核**: `TrendWorker` 定期将 "待复核" 消息发送给 `ai_service` 进行有效性评估。
    - Prompt: "这条消息（发布于X小时前）对当前股价是否仍有显著影响？"
    - Result: 如果 "No"，物理删除或软删除；如果 "Yes"，更新时间戳/权重，继续保留。
  - **长期政策**: 默认保留较长时间，定期复核是否已被新政策覆盖。

### 5. 代码重构
- **Decision**: 拆分 `langchain_provider.go`。
- **Structure**:
  - `backend/ai_service/biz/provider/llm/predictor/`: 存放股价预测逻辑。
  - `backend/ai_service/biz/provider/llm/analyst/`: 存放市场分析逻辑 (含 `ProcessMarketTrends` 实现)。
  - `backend/ai_service/biz/provider/llm/vision/`: 存放图像识别逻辑。

## Alternatives Considered

### 1. 存储在 `stock_service`
- **Option**: 将 `market_trends` 表放在 `stock_service`，`ai_service` 通过 RPC 调用。
- **Pros**: 数据集中管理，`stock_service` 可能是所有市场数据的 source of truth。
- **Cons**: 增加了 RPC 调用开销；`ai_service` 产生的 AI 衍生数据（Intelligence）可以由自己管理。
- **Decision**: 暂时在 `ai_service` 内部实现闭环，因为这更像是 AI 产生的“知识”而非原始市场数据。如果未来其他服务也需要，可以迁移或通过 RPC 暴露。

### 2. 实时抓取 + 缓存
- **Option**: 继续实时抓取，但增加 Redis 缓存。
- **Pros**: 实现简单。
- **Cons**: 无法解决“冷启动”时的慢速问题；无法解决 LLM 处理的高延迟（如果要加 LLM 过滤，实时处理太慢）。

## Risks
- **IP 封禁**: 频繁抓取可能导致 IP 被封。
  - **Mitigation**: 降低抓取频率（如每小时一次），使用代理池（如果需要）。
- **LLM 成本**: 处理大量热搜可能消耗 Tokens。
  - **Mitigation**: 先用关键词规则过滤一遍，再送 LLM；或者仅对 Top N 热搜进行 LLM 处理。

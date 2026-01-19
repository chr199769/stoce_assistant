# Change: Add Smart Trading Enhancement Suite

## Why
目前应用缺乏实时洞察、云端同步和量化工具，与同花顺/开盘啦等竞品相比竞争力不足。为了提升用户体验和专业度，需要引入智能交易增强套件。
此外，为了提升 AI 预测的深度和准确性，需要引入**知识图谱 (Knowledge Graph)** 以理解产业链关系，并引入**预测追踪与评估 (Tracking & Eval)** 机制以持续优化模型。
最后，为了支持内部团队的多人使用场景，需要增加**轻量级用户身份 (Simple Identity)** 以隔离自选股数据。

## What Changes
1. **Cloud Watchlist**: 将自选股从本地存储迁移到后端数据库，支持多设备同步和 AI 自动打标签。
2. **Intraday Market Sentinel**: 实时 AI 分析板块轮动和异常成交量（盘中 vs 盘前/盘后）。
3. **Fractal Prediction**: "历史重演"功能，寻找历史上最相似的 3 个 K 线形态以预测未来走势。
4. **Graph RAG (Knowledge Graph)**:
   - 集成 **Neo4j** 图数据库。
   - 存储 `Stock -> Sector`, `Stock -> Supplier/Customer`, `Stock -> Concept` 关系。
   - 增强 AI Context：在预测时检索关联实体。
5. **Prediction Tracking & Eval**:
   - 集成 **Langfuse**。
   - 记录每次预测的 Prompt/Result (Tracing)。
   - 定时任务自动回填预测准确度 (Auto-Eval)。
6. **Simple User Identity**:
   - 简单的用户登录工具（无需密码/注册）。
   - 仅用于生成/获取 UserID 以区分不同用户的自选股数据。

## Impact
- 受影响的规范: `watchlist`, `ai-review`, `stock-prediction`, `knowledge-graph`, `evaluation`, `identity` (new)
- 受影响的代码:
  - `infrastructure`: 新增 Neo4j 和 Langfuse 服务。
  - `stock_service`: 新增 Watchlist/Signal/User 表，新增 Eval Worker。
  - `ai_service`: 集成 Neo4j Driver 和 Langfuse SDK。
  - `mobile`: 新增登录/切换用户界面，新增自选股 UI 和预测图表。

# Change: Refactor Market Info Trends

## Why
当前的 `MarketInfo` 工具直接抓取实时社交热搜（如微博、百度、知乎），存在两个主要问题：
1. **信噪比低**：包含大量娱乐、社会八卦等非金融内容，干扰 LLM 的市场判断。
2. **性能与稳定性差**：实时抓取速度慢，且依赖的第三方接口经常变动或失效。

## What Changes
- **架构分离**：将热点抓取从实时请求中剥离，改为后台独立 Worker 定时执行。
- **智能过滤**：引入 LLM 对抓取的热点进行预处理，过滤非金融相关内容，提取核心金融情报，并打标关联行业/股票。
- **持久化存储**：将处理后的高价值情报存入数据库。
- **工具升级**：`MarketInfo` 工具改为直接查询数据库中的高价值情报，不再实时抓取。

## Impact
- **受影响的规范**: 新增 `market-intelligence` 能力。
- **受影响的代码**:
    - `backend/ai_service`: 新增 `TrendWorker`，新增数据库模型 `MarketTrend`。
    - `backend/ai_service/biz/tool/market_tool.go`: 修改读取逻辑。

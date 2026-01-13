# Change: 优化股票预测数据

## Why
目前的股票预测功能缺乏足够的数据深度。用户需要更多的历史背景（龙虎榜）和技术指标（筹码峰、买卖盘）来做出更好的决策。此外，缺失行业和市场层面的信息，这对于全面分析至关重要。

## What Changes
- **龙虎榜**: 将数据获取范围从最近 1 个交易日扩展到过去 5 个交易日。
- **技术指标**: 增加对筹码峰（Chip Peak）变化、当前买卖盘（Orders）和挂单量（Pending Order Volumes）的支持。
- **市场与行业信息**: 增加获取行业分类/表现和一般市场情绪/指数数据的能力。

## Impact
- **受影响的规范**: `stock-data` (新能力)
- **受影响的代码**: `backend/ai_service/biz/tool/eastmoney_api.go`, `backend/ai_service/biz/tool/stock_tool.go`

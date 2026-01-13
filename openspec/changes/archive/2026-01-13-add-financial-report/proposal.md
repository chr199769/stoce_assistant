# Change: Add Financial Report Capability

## Why
用户需要基本面数据来辅助投资决策。公司财报提供了关键的财务指标，如收入、净利润和增长率，是基本面分析的核心。

## What Changes
- 在 `stock-data` 能力中新增获取公司财报的需求。
- 支持获取最近几个季度的主要财务指标。

## Impact
- **Specs**: `stock-data`
- **Code**: `stock_service` (新增数据获取逻辑), `gateway` (新增 API 接口)

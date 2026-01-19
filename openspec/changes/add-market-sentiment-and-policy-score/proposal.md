# Change: Add Market Sentiment and Policy Score

## Why
Currently, the Market Analysis provides a qualitative summary but lacks quantitative metrics for market sentiment and policy impact. Users need specific scores to quickly gauge the market atmosphere and policy environment.

## What Changes
- **IDL Update**: Add `sentiment_index` (0-100) and `policy_score` (-5 to +5) to `MarketAnalysisResponse` in `api.thrift` and `ai.thrift`.
- **Logic Implementation**:
  - Implement a deterministic **Market Sentiment Index** calculation based on Limit Up/Down counts and Net Inflow.
  - Implement an LLM-based **Policy Impact Score** calculation based on news analysis and sector alignment.
- **AI Service**: Update `AnalyzeMarket` workflow to compute these scores.

## Impact
- **Specs**: `market-analysis` (New capability)
- **Services**: `ai_service` (primary logic), `gateway` (pass-through).
- **IDL**: `api.thrift`, `ai.thrift`.

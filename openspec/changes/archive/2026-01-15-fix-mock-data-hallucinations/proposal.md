# Change: Fix Mock Data Hallucinations

## Why
Currently, the sentiment provider (`GetLimitUpPool`) falls back to hardcoded mock data (e.g., "Moutai limit-up") when the external API fails. This causes the AI agent to hallucinate and provide misleading market analysis based on fake data. We need to remove this dangerous fallback and ensure the AI handles missing data gracefully.

## What Changes
- **Remove Mock Data**: Delete the hardcoded fallback in `backend/stock_service/biz/provider/sentiment/client.go`.
- **Update AI Prompts**: Modify `ReviewMarket` and `AnalyzeMarket` prompts in `backend/ai_service` to explicitly instruct the AI to report "Data Unavailable" instead of fabricating analysis when input data is empty.

## Impact
- **stock-service**: API behavior change (returns error/empty instead of mock).
- **ai-service**: Prompt engineering update for robustness.

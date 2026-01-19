## 1. IDL & Code Generation
- [ ] 1.1 Update `idl/ai.thrift`: Add `sentiment_index` (double) and `policy_score` (double) to `MarketAnalysisResponse`.
- [ ] 1.2 Update `idl/api.thrift`: Add `sentiment_index` (double) and `policy_score` (double) to `MarketAnalysisResponse`.
- [ ] 1.3 Run `kitex` and `hz` code generation scripts to update Go code.

## 2. AI Service Implementation
- [ ] 2.1 Update `ai_service/biz/provider/llm/provider.go` interface to return scores in `AnalyzeMarket`.
- [ ] 2.2 Implement `CalculateSentimentIndex` function in `ai_service`:
    - Formula: `Base(50) + (LimitUp - LimitDown) * 0.5 + NormalizedNetInflow * 0.1` (clamped 0-100).
- [ ] 2.3 Update `ai_service/biz/provider/llm/langchain_provider.go`:
    - Modify `AnalyzeMarket` prompt to ask LLM for a `policy_score` (-5 to +5) based on news.
    - Parse the score from LLM response.
- [ ] 2.4 Update `ai_service/handler.go`:
    - Call sentiment calculation.
    - Pass both scores to the response.

## 3. Gateway Implementation
- [ ] 3.1 Update `gateway/biz/handler/api/stock_api.go`: Map RPC response fields to HTTP response fields.

## 4. Verification
- [ ] 4.1 Run local test for `AnalyzeMarket` and verify JSON response contains valid scores.

## ADDED Requirements

### Requirement: Market Sentiment Index
The system MUST calculate and return a deterministic Market Sentiment Index (0-100) as part of the market analysis.

#### Scenario: Calculate Sentiment Index
- **WHEN** performing market analysis
- **THEN** calculate the index using the following logic:
  - Start with a Base score of 50.
  - Add `(LimitUpCount - LimitDownCount) * 0.5`.
  - Add `NetInflow (in billions) * 0.1`.
  - Clamp the result between 0 and 100.
  - Return this value as `sentiment_index`.

### Requirement: Policy Impact Score
The system MUST provide a Policy Impact Score (-5 to +5) evaluating the current policy environment's effect on the market.

#### Scenario: Generate Policy Score
- **WHEN** performing market analysis via LLM
- **THEN** the LLM MUST evaluate recent news and sector alignment.
- **AND** assign a score from -5 (Very Negative) to +5 (Very Positive).
- **AND** return this value as `policy_score`.

### Requirement: Enhanced Market Analysis Response
The Market Analysis API response MUST include the new quantitative metrics.

#### Scenario: Return Analysis Data
- **WHEN** `AnalyzeMarket` API is called
- **THEN** the response JSON MUST include:
  - `sentiment_index`: Integer or Float (0-100)
  - `policy_score`: Integer or Float (-5 to +5)
  - Existing fields (summary, risks, opportunities, etc.)

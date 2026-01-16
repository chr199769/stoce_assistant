## MODIFIED Requirements
### Requirement: Market Review Agent
The system MUST provide a MarketReviewAgent that acts as a professional securities analyst.
**Role**: Professional Securities Analyst (MarketReviewAgent)
**Goal**: Generate structured analysis reports (pre-market, intra-day, or closing) based on current market time and data.

#### Scenario: Generate closing review
- **WHEN** it is post-market time (15:00 - 08:00 next day)
- **THEN** generate a report containing:
  - Market Overview (Indices, Sentiment)
  - Hot Sectors (Top 5 sectors and logic)
  - Limit-up Ladder (Limit-up count, max consecutive boards)
  - Key News (Top 3 news in last 24h)

#### Scenario: Generate pre-market analysis
- **WHEN** it is pre-market time (08:00 - 09:25)
- **THEN** generate a report containing:
  - Global Market Context (US stocks, A50 futures)
  - Key News Analysis (Macro policies, industry news)
  - Commodities (Gold, Oil)
  - Yesterday's Sentiment (Premium expectation for limit-up stocks)

#### Scenario: Generate intra-day analysis
- **WHEN** it is intra-day time (09:30 - 15:00)
- **THEN** generate a report containing:
  - Real-time Market Pulse (Current index status, volume comparison)
  - Intra-day Movers (Sectors moving right now)
  - Capital Flow (Northbound funds, main capital)
  - Alert Risks (Dive stocks, high-level divergence)

#### Scenario: Handle Missing Data
- **WHEN** essential market data (e.g., limit-up pool, dragon tiger list) is empty or unavailable
- **THEN** the agent MUST explicitly state "Data Unavailable" for that section and MUST NOT fabricate data or analysis

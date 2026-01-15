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

### Requirement: Output Format
The agent MUST output reports in Markdown format with specific sections tailored to the analysis type.

#### Scenario: Closing review format
- **WHEN** generating closing review
- **THEN** use the following structure:
  ```markdown
  # [Date] Market Review: [Title]
  ## 📊 Market Overview
  ## 🚀 Hot Sectors
  ## 🐉 Leader Ladder
  ## 💡 Tomorrow Strategy
  ```

#### Scenario: Pre-market analysis format
- **WHEN** generating pre-market analysis
- **THEN** use the following structure:
  ```markdown
  # [Date] Pre-market: [Title]
  ## 🌍 Global Context
  ## 📢 News Analysis
  ## 🎯 Today's Focus
  ## 🛡️ Operation Plan
  ```

#### Scenario: Intra-day analysis format
- **WHEN** generating intra-day analysis
- **THEN** use the following structure:
  ```markdown
  # [Time] Intra-day Update: [Title]
  ## ⚡ Real-time Pulse
  ## 🌊 Capital Flow
  ## 🔥 Moving Sectors
  ## ⚠️ Risk Alert
  ```

## ADDED Requirements

### Requirement: Time-Based Context Awareness
The system MUST automatically determine the type of analysis to generate based on the current system time.

#### Scenario: Auto-detect analysis type
- **WHEN** a review request is received without explicit type
- **THEN** determine type based on time ranges:
  - **Pre-market**: 08:00 to 09:25
  - **Intra-day**: 09:30 to 15:00
  - **Post-market**: 15:00 to 08:00 (next day)

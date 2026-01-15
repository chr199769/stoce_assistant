# 规范：AI 市场复盘 (Market Review)

## ADDED Requirements

### Requirement: Market Review Agent
The system MUST provide a MarketReviewAgent that acts as a professional securities analyst.
**Role**: Professional Securities Analyst (MarketReviewAgent)
**Goal**: Generate a structured closing review and pre-market analysis based on market data.

#### Scenario: Generate closing review
- **WHEN** the market closes
- **THEN** generate a report containing:
  - Market Overview (Indices, Sentiment)
  - Hot Sectors (Top 5 sectors and logic)
  - Limit-up Ladder (Limit-up count, max consecutive boards)
  - Key News (Top 3 news in last 24h)

#### Scenario: Generate pre-market analysis
- **WHEN** before market opens
- **THEN** generate a report containing:
  - Global Market Context (US stocks, A50 futures)
  - Key News Analysis (Macro policies, industry news)
  - Commodities (Gold, Oil)
  - Yesterday's Sentiment (Premium expectation for limit-up stocks)

### Requirement: Market Review Toolchain
The agent MUST have access to specific tools to retrieve market data.

#### Scenario: Tool access
- **WHEN** the agent executes
- **THEN** it can call:
  - `GetMarketIndices`: For index data
  - `GetMarketSectors`: For sector rankings
  - `GetLimitUpPool`: For sentiment data
  - `MarketInfoTool`: For news and global market data

### Requirement: Output Format
The agent MUST output reports in Markdown format with specific sections.

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

# 规范：Stock Data API 扩展

## ADDED Requirements

### Requirement: Sector Rotation Data
The system MUST provide real-time sector ranking data including concept and industry sectors.

#### Scenario: Get sector rankings
- **WHEN** a user requests sector rankings
- **THEN** the system returns a list of sectors sorted by change percent or net inflow.
- **AND** the response includes:
  - Sector Code and Name
  - Change Percent
  - Main Net Inflow
  - Top Stock Name
  - Type (Concept/Industry)

### Requirement: Market Sentiment Data
The system MUST provide data on limit-up stocks and market sentiment indicators.

#### Scenario: Get limit-up pool
- **WHEN** a user requests the limit-up pool
- **THEN** the system returns a summary of limit-up counts and a list of stocks.
- **AND** the summary includes:
  - Limit-up count
  - Broken limit-up count
  - Limit-down count
- **AND** the stock list includes:
  - Stock Code and Name
  - Price and Change Percent
  - Limit-up Type (e.g., 2 consecutive boards)
  - Reason (e.g., Huawei Concept)
  - Is Broken status

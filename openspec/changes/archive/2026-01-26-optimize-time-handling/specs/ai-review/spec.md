## MODIFIED Requirements
### Requirement: Time-Based Context Awareness
The system MUST automatically determine the type of analysis and the target date description based on the current system time.

#### Scenario: Auto-detect analysis type and date
- **WHEN** a review request is received without explicit type
- **THEN** determine type based on time ranges:
  - **Pre-market**: 08:00 to 09:25
  - **Intra-day**: 09:30 to 15:00
  - **Post-market**: 15:00 to 08:00 (next day)
- **AND** determine target date description for the prompt:
  - If Pre-market (Today): "今天"
  - If Post-market (Today, before midnight): "明天" (or next trading day)
  - If Post-market (Tomorrow, before pre-market): "今天"

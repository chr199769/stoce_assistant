## ADDED Requirements

### Requirement: Stock Prediction Resonance
系统 MUST 提供基于板块-个股共振的预测分析数据。

#### Scenario: 获取共振预测
- **WHEN** 用户请求股票预测
- **THEN** 系统返回该股票的共振评分（Resonance Score），反映其与所属板块趋势的一致性。
- **AND** 如果板块处于上升趋势且个股为龙头或跟随上涨，评分应较高。

## MODIFIED Requirements

### Requirement: Sector Rotation Data
The system MUST provide real-time sector ranking data including concept and industry sectors, and detailed information for specific sectors.

#### Scenario: Get sector rankings
- **WHEN** a user requests sector rankings
- **THEN** the system returns a list of sectors sorted by change percent or net inflow.
- **AND** the response includes:
  - Sector Code and Name
  - Change Percent
  - Main Net Inflow
  - Top Stock Name
  - Type (Concept/Industry)

#### Scenario: 获取板块详情
- **WHEN** 用户请求指定板块的详情
- **THEN** 系统返回该板块的详细绩效指标和成份股列表。
- **AND** 响应中必须标识出“龙头股”（Leader Stock），基于市场表现（如连板数、涨幅）。

### Requirement: 历史龙虎榜
系统 MUST 提供每日龙虎榜榜单数据以及指定股票的历史上榜记录，并包含席位分析。

#### Scenario: 最近上榜的股票
- **WHEN** 股票在过去 5 天内出现在龙虎榜上
- **THEN** 系统返回这些上榜的日期、原因和净买入/卖出金额

#### Scenario: 获取每日龙虎榜
- **WHEN** 用户请求特定日期的龙虎榜全榜单
- **THEN** 系统返回当天上榜的所有股票列表，包含上榜原因、收盘价、涨跌幅和净买入额。

#### Scenario: 席位游资分析
- **WHEN** 获取龙虎榜席位数据
- **THEN** 系统提供席位名称与知名游资（HotMoney）的映射关系（例如“中信证券西安朱雀大街”映射为“方新侠”）。

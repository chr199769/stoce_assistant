# watchlist Specification

## Purpose
TBD - created by archiving change add-watchlist-deletion. Update Purpose after archive.
## Requirements
### Requirement: Remove Stock from Watchlist
系统 MUST 支持用户从自选股列表中移除股票，并同步至云端。

#### Scenario: Remove existing stock
- **WHEN** 用户点击某只股票的删除按钮
- **THEN** 该股票从界面列表中移除
- **AND** 后端数据库删除该记录（或标记为删除）
- **AND** 变更同步到其他设备

### Requirement: Cloud Watchlist Sync
系统 MUST 支持用户自选股数据的云端存储和多设备同步。

#### Scenario: Sync watchlist across devices
- **WHEN** 用户在设备 A 上添加股票到自选股
- **THEN** 该股票应出现在用户登录的设备 B 的自选股列表中
- **AND** 后端数据库应持久化该记录

### Requirement: AI Auto-Tagging
系统 MUST 自动为自选股中的股票添加 AI 生成的标签（如“近期新高”、“放量上涨”）。

#### Scenario: Auto-tagging stocks
- **WHEN** 用户查看自选股列表
- **THEN** 股票卡片上显示 AI 分析生成的动态标签


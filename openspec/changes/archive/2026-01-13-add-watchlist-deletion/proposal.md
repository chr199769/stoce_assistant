# Change: 支持在移动端删除自选股票

## Why
目前用户只能添加自选股，无法删除，导致列表冗余且无法清理，影响用户体验。

## What Changes
- 在移动端自选股卡片上添加删除功能。
- 更新本地存储逻辑，支持移除已删除的股票代码。

## Impact
- 受影响的规范: watchlist
- 受影响的代码: mobile/src/screens/HomeScreen.tsx

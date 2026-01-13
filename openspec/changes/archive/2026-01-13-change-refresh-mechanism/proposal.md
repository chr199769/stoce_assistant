# Change: 更改移动端数据刷新机制

## Why
用户反馈希望将手动下拉刷新改为自动轮询，以提供更实时的股票价格体验。

## What Changes
- 移除 HomeScreen 的下拉刷新功能
- 实现 3 秒间隔的自动轮询
- 使用 useFocusEffect 确保仅在页面可见时轮询

## Impact
- 受影响的规范: mobile-watchlist
- 受影响的代码: mobile/src/screens/HomeScreen.tsx

## 1. Backend Implementation (AI Service)
- [ ] 1.1 修改 `idl/ai.thrift`，在 `MarketReviewResponse` 中增加 `hot_stocks` (热门股票) 和 `recommended_stocks` (推荐股票) 字段。
- [ ] 1.2 重新生成 Kitex 代码 (`kitex_gen`)。
- [ ] 1.3 更新 `langchain_provider.go` 中的 `ReviewMarket` 方法：
    - [ ] 修改 Prompt，明确要求 AI 输出热门股票和推荐股票列表。
    - [ ] 更新 JSON 解析逻辑，映射新字段。

## 2. Mobile Implementation
- [ ] 2.1 创建 `src/screens/MarketAnalysisScreen.tsx`：
    - [ ] 实现 Tab 切换：`复盘总结` (Market Review) 和 `盘前分析` (Pre-market Analysis)。
    - [ ] `复盘总结` Tab 展示：市场总览、板块分析、情绪分析、游资动向。
    - [ ] `盘前分析` Tab 展示：风险提示、热门股票、推荐股票。
- [ ] 2.2 重构 `src/screens/HomeScreen.tsx`：
    - [ ] 移除跳转到 `PredictionScreen` 的逻辑。
    - [ ] 在股票卡片 (`StockCard`) 中添加 "AI 预测" 按钮。
    - [ ] 实现预测结果的行内展开展示 (显示分析摘要、置信度、新闻摘要)。
- [ ] 2.3 更新 `src/navigation/AppNavigator.tsx`：
    - [ ] 将 `MarketAnalysisScreen` 加入底部导航栏 (Tab Navigator)，替换原有的 `Summary` 和 `PredictionTab`。
    - [ ] 调整 Tab 图标和标题。

## 3. Verification
- [ ] 3.1 验证后端 `MarketReview` 接口能否正确返回新增的股票列表字段。
- [ ] 3.2 验证移动端新页面 `MarketAnalysisScreen` 数据展示正常。
- [ ] 3.3 验证首页点击 "AI 预测" 能正确获取并展示预测结果。
 
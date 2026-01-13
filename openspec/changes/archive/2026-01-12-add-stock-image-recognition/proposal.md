# Change: Add Stock Image Recognition and Mobile Watchlist

## Why
用户希望能够通过导入图片快速添加自选股票，而无需手动输入股票代码或名称。此外，为了提升用户体验，自选股列表需要持久化存储在本地，以便用户下次打开应用时能够直接查看。

## What Changes
- **移动端**:
  - 增加图片选择功能（从相册导入）。
  - 集成后端图片识别接口，展示识别结果。
  - 实现自选股列表的本地存储（使用 `AsyncStorage`）。
- **后端 (Gateway)**:
  - 新增 `POST /api/image/recognize` 接口，接收图片文件。
- **后端 (AI Service)**:
  - 新增 `ImageRecognition` RPC 方法。
  - 集成支持视觉能力的 LLM（如 GPT-4o 或 GLM-4v）进行图片内容分析，提取股票信息。

## Impact
- **Affected specs**: `ai-recognition`, `mobile-watchlist`
- **Affected code**:
  - Mobile: `mobile/src/screens/HomeScreen.tsx`, `mobile/package.json`
  - Gateway: `backend/gateway/idl/api.thrift`, `backend/gateway/biz/handler`
  - AI Service: `backend/ai_service/idl/ai.thrift`, `backend/ai_service/biz/service`
  - Stock Service: 无直接影响，但识别结果可能需要校验股票是否存在（暂不强校验）。

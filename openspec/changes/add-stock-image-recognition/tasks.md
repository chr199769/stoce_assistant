## 1. Implementation

### Backend - IDL & Generation
- [ ] 1.1 Update `backend/idl/ai.thrift` to include `ImageRecognitionRequest` and `ImageRecognitionResponse`.
- [ ] 1.2 Update `backend/idl/api.thrift` to include HTTP mapping for image recognition.
- [ ] 1.3 Run Kitex/Hertz code generation for AI Service and Gateway.

### Backend - AI Service
- [ ] 1.4 Implement `ImageRecognition` method in `ai_service`.
- [ ] 1.5 Integrate LLM vision capability (e.g., using LangChainGo or direct API call if needed).

### Backend - Gateway
- [ ] 1.6 Implement `RecognizeStockImage` handler in `gateway`.
- [ ] 1.7 Verify file upload handling and forwarding to AI Service.

### Mobile - Dependencies & UI
- [ ] 1.8 Install `react-native-image-picker` and configure permissions (Android/iOS).
- [ ] 1.9 Create/Update `StockService` in mobile to call the new API.
- [ ] 1.10 Add "Import Image" button to `HomeScreen`.

### Mobile - Logic & Storage
- [ ] 1.11 Implement image selection and upload logic.
- [ ] 1.12 Handle recognition response and display candidates to user.
- [ ] 1.13 Implement `AsyncStorage` logic for persisting the watchlist.
- [ ] 1.14 Load watchlist from storage on app launch.

### Verification
- [ ] 1.15 Verify end-to-end flow: Image Import -> Recognition -> Add to Watchlist -> Persist -> Relaunch.

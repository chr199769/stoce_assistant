## 1. Implementation
- [x] 1.1 在 `stock_service` 中定义财报数据的 IDL 结构
- [x] 1.2 在 `stock_service` 中实现获取财报数据的 Provider (使用东方财富 API: `https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_LICO_FN_CPD`)
- [x] 1.3 在 `stock_service` 中实现 Handler 逻辑
- [x] 1.4 在 `gateway` 中暴露获取财报的 HTTP 接口
- [x] 1.5 编写单元测试验证数据获取逻辑

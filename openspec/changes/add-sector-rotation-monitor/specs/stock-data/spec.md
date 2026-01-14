# 规范：Stock Data API 扩展

## 1. 数据模型

### 1.1 板块信息 (SectorInfo)
```go
type SectorInfo struct {
    Code        string  `json:"code"`         // 板块代码
    Name        string  `json:"name"`         // 板块名称
    ChangePercent float64 `json:"change_pct"` // 涨跌幅
    MainNetInflow float64 `json:"net_inflow"` // 主力净流入
    TopStockName  string  `json:"top_stock"`  // 领涨股名称
    Type          string  `json:"type"`       // 类型：Concept(概念), Industry(行业)
}
```

### 1.2 涨停股信息 (LimitUpStock)
```go
type LimitUpStock struct {
    Code          string  `json:"code"`
    Name          string  `json:"name"`
    Price         float64 `json:"price"`
    ChangePercent float64 `json:"change_pct"`
    LimitUpType   string  `json:"limit_up_type"` // 首板, 2连板, 3连板...
    Reason        string  `json:"reason"`        // 涨停原因 (如：华为概念+芯片)
    IsBroken      bool    `json:"is_broken"`     // 是否炸板
}
```

## 2. 接口定义

### 2.1 获取板块排行
`GET /market/sectors?type={industry|concept}&sort={change|inflow}`

**Response:**
```json
{
  "sectors": [
    {
      "name": "半导体",
      "change_pct": 5.2,
      "top_stock": "中芯国际"
    },
    ...
  ]
}
```

### 2.2 获取涨停数据
`GET /market/limit_up_pool`

**Response:**
```json
{
  "summary": {
    "limit_up_count": 45,
    "broken_count": 12,
    "limit_down_count": 2
  },
  "limit_up_list": [...]
}
```

# AkShare HTTP Service

## Endpoints
### GET /macro
- Query:
  - funcs: comma-separated AkShare function names (default: macro_china_cpi,macro_china_ppi,macro_china_pmi,macro_china_gdp,macro_china_fx_reserves,macro_china_money_supply)
  - limit: integer, tail records count per function (default: 12)
  - symbol: optional symbol for certain functions
- Response: JSON array of items
  - indicator, label, value, unit, time, source

### GET /news
- Query:
  - func: primary AkShare function name (default: fallback stock_news_em or news_cctv)
  - symbol: optional, used by stock_news_em (default: A股 when empty)
  - limit: integer (default: 50)
- Response: JSON array of items
  - title, content, url, source, time

## Configuration
- Environment:
  - AK_PYTHON_BIN: Python executable (default: python3)
  - AK_SERVICE_ADDR: service listen address (default: :8895)

## Integration
- Gateway proxies:
  - GET /api/akshare/macro -> queries AkShare HTTP service
  - GET /api/akshare/news -> queries AkShare HTTP service
- Stock Service:
  - Uses akshare.service_url for priority HTTP call
  - Falls back to local Python execution when HTTP fails

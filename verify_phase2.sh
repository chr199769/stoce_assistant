#!/bin/bash

echo "Testing GetSectorStocks (BK0475)..."
curl -s "http://localhost:8080/api/stock/sector/stocks?sector_code=BK0475" | head -c 500
echo "..."
echo ""

echo "Testing GetDragonTigerList (2026-01-14)..."
curl -s "http://localhost:8080/api/stock/dragontiger/list?date=2026-01-14" | head -c 1000
echo "..."
echo ""

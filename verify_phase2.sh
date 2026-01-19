#!/bin/bash

BASE_URL="http://localhost:8080/api"

echo "=== Testing Phase 2 APIs ==="

echo "1. Market Sectors (Concept)"
curl -s "$BASE_URL/market/sectors?type=concept&limit=5" | jq .

echo -e "\n\n2. Limit Up Pool"
curl -s "$BASE_URL/market/limit_up" | jq .

echo -e "\n\n3. Login/Create User (TestUser)"
USER_RESP=$(curl -s -X POST "$BASE_URL/user/login" -H "Content-Type: application/json" -d '{"username":"TestUser"}')
echo $USER_RESP | jq .
USER_ID=$(echo $USER_RESP | jq -r '.user.id')

if [ "$USER_ID" == "null" ]; then
    echo "Login failed!"
    exit 1
fi

echo -e "\nUser ID: $USER_ID"

echo -e "\n\n4. Add Watchlist (600519)"
curl -s -X POST "$BASE_URL/watchlist/add" -H "Content-Type: application/json" -d "{\"user_id\":\"$USER_ID\", \"stock_code\":\"600519\"}" | jq .

echo -e "\n\n5. Get Watchlist"
curl -s "$BASE_URL/watchlist/list?user_id=$USER_ID" | jq .

echo -e "\n\n6. Remove Watchlist (600519)"
curl -s -X POST "$BASE_URL/watchlist/remove" -H "Content-Type: application/json" -d "{\"user_id\":\"$USER_ID\", \"stock_code\":\"600519\"}" | jq .

echo -e "\n\n7. Historical Kline (sh600519, 5 days)"
curl -s "$BASE_URL/stock/kline?code=600519&days=5" | jq .

echo -e "\n\n8. Dragon Tiger List"
curl -s "$BASE_URL/stock/dragontiger/list" | jq 'del(.items[].buy_seats) | del(.items[].sell_seats)' # Truncate seats for display

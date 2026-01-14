#!/bin/bash
# start_services.sh

# Kill existing processes
pkill -f "stock_service"
pkill -f "ai_service"
pkill -f "gateway"

mkdir -p logs

echo "Starting Stock Service..."
cd backend/stock_service && go run . > ../../logs/stock_service.log 2>&1 &
STOCK_PID=$!
echo "Stock Service PID: $STOCK_PID"

echo "Starting AI Service..."
cd backend/ai_service && go run . > ../../logs/ai_service.log 2>&1 &
AI_PID=$!
echo "AI Service PID: $AI_PID"

echo "Starting Gateway..."
cd backend/gateway && go run . > ../../logs/gateway.log 2>&1 &
GATEWAY_PID=$!
echo "Gateway PID: $GATEWAY_PID"

echo "Waiting for services to initialize (5s)..."
sleep 5

echo "Services running."
echo "Stock Service (PID: $STOCK_PID) - Port 8888"
echo "AI Service (PID: $AI_PID) - Port 8889"
echo "Gateway (PID: $GATEWAY_PID) - Port 8080"
echo "Check logs in logs/stock_service.log, logs/ai_service.log, logs/gateway.log"

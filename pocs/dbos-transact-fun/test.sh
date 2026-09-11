#!/bin/bash

set -e

echo "Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/health)
echo "Health response: $HEALTH_RESPONSE"

if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    echo "Health check passed"
else
    echo "Health check failed"
    exit 1
fi

echo ""
echo "Testing order endpoint..."
ORDER_RESPONSE=$(curl -s -X POST http://localhost:8080/order \
    -H "Content-Type: application/json" \
    -d '{"item":"laptop","quantity":2}')

echo "Order response: $ORDER_RESPONSE"

if echo "$ORDER_RESPONSE" | grep -q "order_id"; then
    echo "Order test passed"
else
    echo "Order test failed"
    exit 1
fi

echo ""
echo "Testing another order..."
ORDER_RESPONSE_2=$(curl -s -X POST http://localhost:8080/order \
    -H "Content-Type: application/json" \
    -d '{"item":"phone","quantity":5}')

echo "Second order response: $ORDER_RESPONSE_2"

if echo "$ORDER_RESPONSE_2" | grep -q "order_id"; then
    echo "Second order test passed"
else
    echo "Second order test failed"
    exit 1
fi

echo ""
echo "All tests passed successfully"

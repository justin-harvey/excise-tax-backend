#!/bin/bash

# Test script for the tax API endpoints

BASE_URL="http://localhost:8081"

echo "🧪 Testing Tax API Endpoints"
echo "============================="

# Start the API server in the background
echo "Starting API server..."
cd /Users/maxfelker/excise-tax-backend/v2
PORT=8081 XRPL_URL=wss://s.altnet.rippletest.net:51233 go run ./cmd/api/ > /tmp/api.log 2>&1 &
API_PID=$!

# Wait for server to start
sleep 3

# Test 1: Health check
echo
echo "1. Testing health endpoint..."
curl -s "$BASE_URL/health" | jq '.'

# Test 2: Info endpoint (should show tax capabilities)
echo
echo "2. Testing info endpoint..."
curl -s "$BASE_URL/info" | jq '.'

# Test 3: Get product types
echo
echo "3. Testing product types endpoint..."
curl -s "$BASE_URL/tax/product-types" | jq '.'

# Test 4: Get tax rates
echo
echo "4. Testing tax rates endpoint..."
curl -s "$BASE_URL/tax/rates" | jq '.'

# Test 5: Get tax rates filtered by jurisdiction
echo
echo "5. Testing tax rates endpoint with jurisdiction filter..."
curl -s "$BASE_URL/tax/rates?jurisdiction=federal" | jq '.'

# Test 6: Validate production data
echo
echo "6. Testing validation endpoint..."
curl -s -X POST "$BASE_URL/tax/validate" \
  -H "Content-Type: application/json" \
  -d '{
    "production_data": {
      "items": [
        {
          "product_type": "beer",
          "product_name": "Test IPA",
          "unit_type": "gallon",
          "quantity": 100,
          "abv": 6.5
        }
      ]
    }
  }' | jq '.'

# Test 7: Calculate tax
echo
echo "7. Testing tax calculation endpoint..."
curl -s -X POST "$BASE_URL/tax/calculate" \
  -H "Content-Type: application/json" \
  -d '{
    "production_data": {
      "items": [
        {
          "product_type": "beer",
          "product_name": "Test IPA",
          "unit_type": "gallon",
          "quantity": 100,
          "abv": 6.5,
          "notes": "Seasonal IPA batch"
        },
        {
          "product_type": "wine",
          "product_name": "Test Chardonnay",
          "unit_type": "gallon",
          "quantity": 50,
          "abv": 12.5
        }
      ]
    },
    "jurisdiction": "federal"
  }' | jq '.'

# Test 8: Test invalid data validation
echo
echo "8. Testing validation with invalid data..."
curl -s -X POST "$BASE_URL/tax/validate" \
  -H "Content-Type: application/json" \
  -d '{
    "production_data": {
      "items": [
        {
          "product_type": "invalid_type",
          "product_name": "",
          "unit_type": "gallon",
          "quantity": -5,
          "abv": 150
        }
      ]
    }
  }' | jq '.'

# Cleanup
echo
echo "Stopping API server..."
kill $API_PID 2>/dev/null
wait $API_PID 2>/dev/null

echo "✅ Tax API testing complete!"
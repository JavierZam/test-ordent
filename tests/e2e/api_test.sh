#!/bin/bash

# E2E Test Script for E-Commerce API
# This script tests the main API endpoints using curl

# Configuration
API_URL="http://localhost:8080/api"
ADMIN_USERNAME="admin_test"
ADMIN_PASSWORD="Admin123"
CUSTOMER_USERNAME="customer_test"
CUSTOMER_PASSWORD="Customer123"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to test endpoints
test_endpoint() {
    local endpoint=$1
    local method=$2
    local expected_status=$3
    local data=$4
    local token=$5
    local description=$6
    
    echo -e "\n=== Testing $description ($method $endpoint) ==="
    
    # Set up curl command based on method and data
    if [ "$method" == "GET" ]; then
        if [ -z "$token" ]; then
            response=$(curl -s -o response.txt -w "%{http_code}" -X $method $API_URL$endpoint)
        else
            response=$(curl -s -o response.txt -w "%{http_code}" -X $method -H "Authorization: Bearer $token" $API_URL$endpoint)
        fi
    else
        if [ -z "$token" ]; then
            response=$(curl -s -o response.txt -w "%{http_code}" -X $method -H "Content-Type: application/json" -d "$data" $API_URL$endpoint)
        else
            response=$(curl -s -o response.txt -w "%{http_code}" -X $method -H "Content-Type: application/json" -H "Authorization: Bearer $token" -d "$data" $API_URL$endpoint)
        fi
    fi
    
    # Check status code
    if [ "$response" -eq "$expected_status" ]; then
        echo -e "${GREEN}✓ Status code: $response (Expected: $expected_status)${NC}"
        cat response.txt | jq . 2>/dev/null || cat response.txt
        return 0
    else
        echo -e "${RED}✗ Status code: $response (Expected: $expected_status)${NC}"
        cat response.txt | jq . 2>/dev/null || cat response.txt
        return 1
    fi
}

# Start Testing
echo "=== Starting E2E Tests for E-Commerce API ==="

# 1. Register admin user
test_endpoint "/auth/admin-register" "POST" 201 '{"username":"'$ADMIN_USERNAME'","email":"admin@test.com","password":"'$ADMIN_PASSWORD'","full_name":"Admin Test","admin_secret":"thisisaverysecretkey"}' "" "Admin Registration"

# 2. Login as admin
test_endpoint "/auth/login" "POST" 200 '{"username":"'$ADMIN_USERNAME'","password":"'$ADMIN_PASSWORD'"}' "" "Admin Login"
ADMIN_TOKEN=$(cat response.txt | jq -r .token)
echo "Admin token: $ADMIN_TOKEN"

# 3. Register customer
test_endpoint "/auth/register" "POST" 201 '{"username":"'$CUSTOMER_USERNAME'","email":"customer@test.com","password":"'$CUSTOMER_PASSWORD'","full_name":"Customer Test"}' "" "Customer Registration"

# 4. Login as customer
test_endpoint "/auth/login" "POST" 200 '{"username":"'$CUSTOMER_USERNAME'","password":"'$CUSTOMER_PASSWORD'"}' "" "Customer Login"
CUSTOMER_TOKEN=$(cat response.txt | jq -r .token)
echo "Customer token: $CUSTOMER_TOKEN"

# 5. Create product (admin)
test_endpoint "/products" "POST" 201 '{"name":"Test Product","description":"A test product","price":99.99,"stock":100,"category_id":1}' "$ADMIN_TOKEN" "Create Product"
PRODUCT_ID=$(cat response.txt | jq -r .id)
echo "Created Product ID: $PRODUCT_ID"

# 6. Get product (public)
test_endpoint "/products/$PRODUCT_ID" "GET" 200 "" "" "Get Product"

# 7. Get customer's cart (should be empty)
test_endpoint "/cart" "GET" 200 "" "$CUSTOMER_TOKEN" "Get Empty Cart"

# 8. Add product to cart
test_endpoint "/cart/items" "POST" 200 '{"product_id":'$PRODUCT_ID',"quantity":2}' "$CUSTOMER_TOKEN" "Add to Cart"
CART_ITEM_ID=$(cat response.txt | jq -r '.items[0].id')
echo "Created Cart Item ID: $CART_ITEM_ID"

# 9. Create order
test_endpoint "/orders" "POST" 201 '{"shipping_address":"123 Test St, Test City"}' "$CUSTOMER_TOKEN" "Create Order"
ORDER_ID=$(cat response.txt | jq -r .id)
echo "Created Order ID: $ORDER_ID"

# 10. Get orders
test_endpoint "/orders" "GET" 200 "" "$CUSTOMER_TOKEN" "Get Orders"

# 11. Try to access admin endpoint as customer (should fail)
test_endpoint "/products" "POST" 403 '{"name":"Unauthorized Product","description":"This should fail","price":9.99,"stock":10,"category_id":1}' "$CUSTOMER_TOKEN" "Unauthorized Product Creation"

# 12. Clean up: Delete the test product
test_endpoint "/products/$PRODUCT_ID" "DELETE" 200 "" "$ADMIN_TOKEN" "Delete Product"

# Cleanup
rm -f response.txt

echo -e "\n=== E2E Tests Completed ==="
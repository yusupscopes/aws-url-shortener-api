#!/bin/bash

# Integration test script for URL Shortener API
# Usage: ./integration_test.sh <api_url>

API_URL=$1

if [ -z "$API_URL" ]; then
  echo "Usage: ./integration_test.sh <api_url>"
  exit 1
fi

echo "Running integration tests against $API_URL"

# Test 1: Create a short URL
echo "Test 1: Create a short URL"
RESPONSE=$(curl -s -X POST "$API_URL/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/test", "expire_in_days": 7}')

echo "Response: $RESPONSE"

# Extract short URL from response
SHORT_URL=$(echo $RESPONSE | grep -o '"short_url":"[^"]*"' | cut -d'"' -f4)

if [ -z "$SHORT_URL" ]; then
  echo "Failed to create short URL"
  exit 1
fi

echo "Created short URL: $SHORT_URL"

# Extract short code from URL
SHORT_CODE=$(echo $SHORT_URL | awk -F'/' '{print $NF}')

echo "Short code: $SHORT_CODE"

# Test 2: Get URL stats
echo "Test 2: Get URL stats"
STATS_RESPONSE=$(curl -s -X GET "$API_URL/stats/$SHORT_CODE")

echo "Stats response: $STATS_RESPONSE"

# Check if stats response contains the original URL
if ! echo $STATS_RESPONSE | grep -q "example.com/test"; then
  echo "Failed to get URL stats"
  exit 1
fi

# Test 3: Redirect to original URL
echo "Test 3: Redirect to original URL"
REDIRECT_RESPONSE=$(curl -s -I -X GET "$SHORT_URL")

echo "Redirect response: $REDIRECT_RESPONSE"

# Check if redirect response contains a 302 Found status
if ! echo $REDIRECT_RESPONSE | grep -q "302 Found"; then
  echo "Failed to redirect to original URL"
  exit 1
fi

# Check if redirect response contains the correct Location header
if ! echo $REDIRECT_RESPONSE | grep -q "Location: https://example.com/test"; then
  echo "Incorrect redirect location"
  exit 1
fi

echo "All tests passed!"
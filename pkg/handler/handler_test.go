package handler

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/database"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/model"
)

func TestShortenURL(t *testing.T) {
	// Setup mock database
	mockDB := database.NewMockDynamoDB().(*database.MockDynamoDB)
	handler := NewHandler(mockDB)

	// Create a mock request
	req := events.LambdaFunctionURLRequest{
		Body: `{"url": "https://example.com", "expire_in_days": 7}`,
		RequestContext: events.LambdaFunctionURLRequestContext{
			DomainName: "test.lambda-url.us-east-1.amazonaws.com",
			RequestID:  "test-request-id",
		},
	}

	// Call the handler
	resp, err := handler.ShortenURL(context.Background(), req)

	// Check for errors
	if err != nil {
		t.Fatalf("ShortenURL returned an error: %v", err)
	}

	// Check status code
	if resp.StatusCode != 201 {
		t.Errorf("Expected status code 201, got %d", resp.StatusCode)
	}

	// Parse response
	var shortenResp model.ShortenResponse
	err = json.Unmarshal([]byte(resp.Body), &shortenResp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check that short URL is not empty
	if shortenResp.ShortURL == "" {
		t.Errorf("Expected non-empty short URL")
	}

	// Check that short URL contains the domain name
	if !strings.Contains(shortenResp.ShortURL, "test.lambda-url.us-east-1.amazonaws.com") {
		t.Errorf("Expected short URL to contain domain name, got %s", shortenResp.ShortURL)
	}
	
	// Test error case - database failure
	mockDB.SetFailNext(true)
	resp, err = handler.ShortenURL(context.Background(), req)
	if err != nil {
		t.Fatalf("ShortenURL should handle errors internally: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("Expected status code 500 on DB error, got %d", resp.StatusCode)
	}
	
	// Test invalid JSON
	badReq := events.LambdaFunctionURLRequest{
		Body: `{"url": "https://example.com", "expire_in_days": }`, // Invalid JSON
		RequestContext: events.LambdaFunctionURLRequestContext{
			DomainName: "test.lambda-url.us-east-1.amazonaws.com",
		},
	}
	resp, err = handler.ShortenURL(context.Background(), badReq)
	if err != nil {
		t.Fatalf("ShortenURL should handle errors internally: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("Expected status code 400 on invalid JSON, got %d", resp.StatusCode)
	}
	
	// Test missing URL
	missingURLReq := events.LambdaFunctionURLRequest{
		Body: `{"expire_in_days": 7}`, // Missing URL
		RequestContext: events.LambdaFunctionURLRequestContext{
			DomainName: "test.lambda-url.us-east-1.amazonaws.com",
		},
	}
	resp, err = handler.ShortenURL(context.Background(), missingURLReq)
	if err != nil {
		t.Fatalf("ShortenURL should handle errors internally: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("Expected status code 400 on missing URL, got %d", resp.StatusCode)
	}
}

func TestRedirectURL(t *testing.T) {
	// Setup mock database
	mockDB := database.NewMockDynamoDB().(*database.MockDynamoDB)
	handler := NewHandler(mockDB)
	
	// Create a test URL in the mock database
	testCode := "testcode"
	mockDB.CreateURL(context.Background(), &model.URLItem{
		ShortCode:   testCode,
		OriginalURL: "https://example.com",
		CreatedAt:   "1234567890",
		Expiration:  9876543210,
		ClickCount:  0,
	})
	
	// Create a mock request
	req := events.LambdaFunctionURLRequest{
		RawPath: "/" + testCode,
		RequestContext: events.LambdaFunctionURLRequestContext{
			DomainName: "test.lambda-url.us-east-1.amazonaws.com",
		},
	}
	
	// Call the handler
	resp, err := handler.RedirectURL(context.Background(), req)
	
	// Check for errors
	if err != nil {
		t.Fatalf("RedirectURL returned an error: %v", err)
	}
	
	// Check status code
	if resp.StatusCode != 302 {
		t.Errorf("Expected status code 302, got %d", resp.StatusCode)
	}
	
	// Check Location header
	location := resp.Headers["Location"]
	if location != "https://example.com" {
		t.Errorf("Expected Location header to be 'https://example.com', got '%s'", location)
	}
	
	// Check click count was incremented
	// Note: We need to wait a bit for the goroutine to complete
	time.Sleep(100 * time.Millisecond)
	urlItem, _ := mockDB.GetURL(context.Background(), testCode)
	if urlItem.ClickCount != 1 {
		t.Errorf("Expected click count to be 1, got %d", urlItem.ClickCount)
	}
	
	// Test empty path
	emptyReq := events.LambdaFunctionURLRequest{
		RawPath: "/",
		RequestContext: events.LambdaFunctionURLRequestContext{
			DomainName: "test.lambda-url.us-east-1.amazonaws.com",
		},
	}
	resp, err = handler.RedirectURL(context.Background(), emptyReq)
	if err != nil {
		t.Fatalf("RedirectURL should handle errors internally: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("Expected status code 400 on empty path, got %d", resp.StatusCode)
	}
	
	// Test non-existent code
	nonExistentReq := events.LambdaFunctionURLRequest{
		RawPath: "/nonexistent",
		RequestContext: events.LambdaFunctionURLRequestContext{
			DomainName: "test.lambda-url.us-east-1.amazonaws.com",
		},
	}
	resp, err = handler.RedirectURL(context.Background(), nonExistentReq)
	if err != nil {
		t.Fatalf("RedirectURL should handle errors internally: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expected status code 404 for non-existent code, got %d", resp.StatusCode)
	}
	
	// Test database error
	mockDB.SetFailNext(true)
	resp, err = handler.RedirectURL(context.Background(), req)
	if err != nil {
		t.Fatalf("RedirectURL should handle errors internally: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("Expected status code 500 on DB error, got %d", resp.StatusCode)
	}
}

func TestGetURLStats(t *testing.T) {
	// Setup mock database
	mockDB := database.NewMockDynamoDB().(*database.MockDynamoDB)
	handler := NewHandler(mockDB)
	
	// Create a test URL in the mock database
	testCode := "testcode"
	mockDB.CreateURL(context.Background(), &model.URLItem{
		ShortCode:   testCode,
		OriginalURL: "https://example.com",
		CreatedAt:   "1234567890",
		Expiration:  9876543210,
		ClickCount:  42,
	})
	
	// Create a mock request
	req := events.LambdaFunctionURLRequest{
		RawPath: "/stats/" + testCode,
		RequestContext: events.LambdaFunctionURLRequestContext{
			DomainName: "test.lambda-url.us-east-1.amazonaws.com",
		},
	}
	
	// Call the handler
	resp, err := handler.GetURLStats(context.Background(), req)
	
	// Check for errors
	if err != nil {
		t.Fatalf("GetURLStats returned an error: %v", err)
	}
	
	// Check status code
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
	
	// Parse response
	var statsResp model.StatsResponse
	err = json.Unmarshal([]byte(resp.Body), &statsResp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	// Check stats
	if statsResp.OriginalURL != "https://example.com" {
		t.Errorf("Expected original URL 'https://example.com', got '%s'", statsResp.OriginalURL)
	}
	if statsResp.ClickCount != 42 {
		t.Errorf("Expected click count 42, got %d", statsResp.ClickCount)
	}
	
	// Test non-existent code
	nonExistentReq := events.LambdaFunctionURLRequest{
		RawPath: "/stats/nonexistent",
		RequestContext: events.LambdaFunctionURLRequestContext{
			DomainName: "test.lambda-url.us-east-1.amazonaws.com",
		},
	}
	resp, err = handler.GetURLStats(context.Background(), nonExistentReq)
	if err != nil {
		t.Fatalf("GetURLStats should handle errors internally: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expected status code 404 for non-existent code, got %d", resp.StatusCode)
	}
	
	// Test database error
	mockDB.SetFailNext(true)
	resp, err = handler.GetURLStats(context.Background(), req)
	if err != nil {
		t.Fatalf("GetURLStats should handle errors internally: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("Expected status code 500 on DB error, got %d", resp.StatusCode)
	}
	
	// Test invalid path
	invalidReq := events.LambdaFunctionURLRequest{
		RawPath: "/stats/", // Missing code
		RequestContext: events.LambdaFunctionURLRequestContext{
			DomainName: "test.lambda-url.us-east-1.amazonaws.com",
		},
	}
	resp, err = handler.GetURLStats(context.Background(), invalidReq)
	if err != nil {
		t.Fatalf("GetURLStats should handle errors internally: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("Expected status code 400 on invalid path, got %d", resp.StatusCode)
	}
}
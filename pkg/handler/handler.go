package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/database"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/model"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/utils"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/logger"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/monitoring"
)

const (
	// Length of the generated short code
	codeLength = 5
)

// ShortenURL handles the creation of a new short URL
func ShortenURL(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	startTime := time.Now()
	logger.Info("Processing shorten URL request", map[string]interface{}{
		"requestId": req.RequestContext.RequestID,
	})

	// Initialize monitoring client
	metricClient, err := monitoring.NewClient(ctx)
	if err != nil {
		logger.Warn("Failed to initialize monitoring client", err)
		// Continue without monitoring
	}

	// Initialize DynamoDB client
	client, err := database.GetClient(ctx)
	if err != nil {
		logger.Error("Failed to initialize DynamoDB client", err)
		if metricClient != nil {
			metricClient.RecordDynamoDBError(ctx, "GetClient")
		}
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"error": "Internal server error: %v"}`, err),
		}, nil
	}

	// Parse request body
	var shortenReq model.ShortenRequest
	err = json.Unmarshal([]byte(req.Body), &shortenReq)
	if err != nil {
		logger.Warn("Invalid request body", map[string]interface{}{
			"body":  req.Body,
			"error": err.Error(),
		})
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "Invalid request body"}`,
		}, nil
	}

	if shortenReq.URL == "" {
		logger.Warn("URL is required but was empty")
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "URL is required"}`,
		}, nil
	}

	// Generate a random code for the short URL
	code, err := utils.GenerateShortCode(codeLength)
	if err != nil {
		logger.Error("Failed to generate short code", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"error": "Failed to generate short code: %v"}`, err),
		}, nil
	}

	// Calculate expiration time if provided
	expiration := utils.CalculateExpirationTime(shortenReq.ExpireInDays)

	// Create URL item
	urlItem := &model.URLItem{
		ShortCode:   code,
		OriginalURL: shortenReq.URL,
		CreatedAt:   time.Now().Format(time.RFC3339),
		Expiration:  expiration,
		ClickCount:  0,
	}

	// Save to DynamoDB
	err = database.CreateURL(ctx, client, urlItem)
	if err != nil {
		logger.Error("Failed to create URL in DynamoDB", map[string]interface{}{
			"shortCode": code,
			"url":       shortenReq.URL,
			"error":     err.Error(),
		})
		if metricClient != nil {
			metricClient.RecordDynamoDBError(ctx, "CreateURL")
		}
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"error": "Failed to create short URL: %v"}`, err),
		}, nil
	}

	// Get base URL from environment variable or use a default
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		// Extract base URL from the request
		baseURL = fmt.Sprintf("https://%s", req.RequestContext.DomainName)
	}

	shortURL := fmt.Sprintf("%s/%s", baseURL, urlItem.ShortCode)
	logger.Info("Successfully created short URL", map[string]interface{}{
		"shortCode":   urlItem.ShortCode,
		"originalURL": urlItem.OriginalURL,
		"shortURL":    shortURL,
		"expiration":  urlItem.Expiration,
	})

	// Record metrics
	if metricClient != nil {
		metricClient.RecordURLCreated(ctx)
		latencyMs := float64(time.Since(startTime).Milliseconds())
		metricClient.RecordAPILatency(ctx, "/shorten", latencyMs)
	}

	response := model.ShortenResponse{
		ShortURL: shortURL,
	}

	responseJSON, _ := json.Marshal(response)
	return events.LambdaFunctionURLResponse{
		StatusCode: http.StatusCreated,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJSON),
	}, nil
}

// RedirectURL handles the redirection to the original URL
func RedirectURL(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	startTime := time.Now()
	
	// Extract code from path
	path := req.RawPath
	if path == "/" {
		logger.Warn("Redirect request with empty path")
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "Short code is required"}`,
		}, nil
	}

	// Initialize monitoring client
	metricClient, err := monitoring.NewClient(ctx)
	if err != nil {
		logger.Warn("Failed to initialize monitoring client", err)
		// Continue without monitoring
	}

	// Remove leading slash
	code := strings.TrimPrefix(path, "/")
	logger.Info("Processing redirect request", map[string]interface{}{
		"shortCode": code,
		"requestId": req.RequestContext.RequestID,
	})

	// Initialize DynamoDB client
	client, err := database.GetClient(ctx)
	if err != nil {
		logger.Error("Failed to initialize DynamoDB client", err)
		if metricClient != nil {
			metricClient.RecordDynamoDBError(ctx, "GetClient")
		}
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"error": "Internal server error: %v"}`, err),
		}, nil
	}

	// Get URL from DynamoDB
	urlItem, err := database.GetURL(ctx, client, code)
	if err != nil {
		if strings.Contains(err.Error(), "URL not found") {
			logger.Warn("URL not found for code", map[string]interface{}{
				"shortCode": code,
			})
			if metricClient != nil {
				metricClient.RecordURLNotFound(ctx)
			}
			return events.LambdaFunctionURLResponse{
				StatusCode: http.StatusNotFound,
				Body:       `{"error": "URL not found"}`,
			}, nil
		}
		
		logger.Error("Failed to retrieve URL from DynamoDB", map[string]interface{}{
			"shortCode": code,
			"error":     err.Error(),
		})
		if metricClient != nil {
			metricClient.RecordDynamoDBError(ctx, "GetURL")
		}
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"error": "Failed to retrieve URL: %v"}`, err),
		}, nil
	}

	// Increment click count (don't wait for the result)
	go func() {
		err := database.IncrementClickCount(context.Background(), client, code)
		if err != nil {
			logger.Error("Failed to increment click count", map[string]interface{}{
				"shortCode": code,
				"error":     err.Error(),
			})
			if metricClient != nil {
				metricClient.RecordDynamoDBError(ctx, "IncrementClickCount")
			}
		}
	}()

	logger.Info("Redirecting to original URL", map[string]interface{}{
		"shortCode":   code,
		"originalURL": urlItem.OriginalURL,
		"clickCount":  urlItem.ClickCount + 1, // +1 because we're incrementing
	})

	// Record metrics
	if metricClient != nil {
		metricClient.RecordURLRedirected(ctx)
		latencyMs := float64(time.Since(startTime).Milliseconds())
		metricClient.RecordAPILatency(ctx, "/{shortCode}", latencyMs)
	}

	// Redirect to the original URL
	return events.LambdaFunctionURLResponse{
		StatusCode: http.StatusFound,
		Headers: map[string]string{
			"Location": urlItem.OriginalURL,
		},
		Body: "",
	}, nil
}

// GetURLStats retrieves analytics for a short URL
func GetURLStats(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	startTime := time.Now()
	
	// Extract code from path
	path := req.RawPath
	code := strings.TrimPrefix(path, "/stats/")
	if code == "" || code == path {
		logger.Warn("Stats request with invalid path", map[string]interface{}{
			"path": path,
		})
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "Short code is required"}`,
		}, nil
	}

	// Initialize monitoring client
	metricClient, err := monitoring.NewClient(ctx)
	if err != nil {
		logger.Warn("Failed to initialize monitoring client", err)
		// Continue without monitoring
	}

	logger.Info("Processing stats request", map[string]interface{}{
		"shortCode": code,
		"requestId": req.RequestContext.RequestID,
	})

	// Initialize DynamoDB client
	client, err := database.GetClient(ctx)
	if err != nil {
		logger.Error("Failed to initialize DynamoDB client", err)
		if metricClient != nil {
			metricClient.RecordDynamoDBError(ctx, "GetClient")
		}
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"error": "Internal server error: %v"}`, err),
		}, nil
	}

	// Get URL from DynamoDB
	urlItem, err := database.GetURL(ctx, client, code)
	if err != nil {
		if strings.Contains(err.Error(), "URL not found") {
			logger.Warn("URL not found for stats", map[string]interface{}{
				"shortCode": code,
			})
			if metricClient != nil {
				metricClient.RecordURLNotFound(ctx)
			}
			return events.LambdaFunctionURLResponse{
				StatusCode: http.StatusNotFound,
				Body:       `{"error": "URL not found"}`,
			}, nil
		}
		
		logger.Error("Failed to retrieve URL for stats", map[string]interface{}{
			"shortCode": code,
			"error":     err.Error(),
		})
		if metricClient != nil {
			metricClient.RecordDynamoDBError(ctx, "GetURL")
		}
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"error": "Failed to retrieve URL: %v"}`, err),
		}, nil
	}

	// Create stats response
	stats := model.StatsResponse{
		OriginalURL: urlItem.OriginalURL,
		CreatedAt:   urlItem.CreatedAt,
		Expiration:  urlItem.Expiration,
		ClickCount:  urlItem.ClickCount,
	}

	logger.Info("Retrieved stats for URL", map[string]interface{}{
		"shortCode":   code,
		"originalURL": urlItem.OriginalURL,
		"clickCount":  urlItem.ClickCount,
		"createdAt":   urlItem.CreatedAt,
		"expiration":  urlItem.Expiration,
	})

	// Record metrics
	if metricClient != nil {
		metricClient.RecordURLStatsRetrieved(ctx)
		latencyMs := float64(time.Since(startTime).Milliseconds())
		metricClient.RecordAPILatency(ctx, "/stats/{shortCode}", latencyMs)
	}

	responseJSON, _ := json.Marshal(stats)
	return events.LambdaFunctionURLResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJSON),
	}, nil
}
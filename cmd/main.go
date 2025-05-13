package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/database"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/handler"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/logger"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/monitoring"
)

func router(ctx context.Context, event events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	startTime := time.Now()
	path := event.RawPath
	method := event.RequestContext.HTTP.Method

	logger.Info("Received request", map[string]interface{}{
		"method":    method,
		"path":      path,
		"requestId": event.RequestContext.RequestID,
		"source":    event.RequestContext.HTTP.SourceIP,
	})

	// Initialize monitoring client
	metricClient, err := monitoring.NewClient(ctx)
	if err != nil {
		logger.Warn("Failed to initialize monitoring client", err)
		// Continue without monitoring
	}

	// Create database interface directly
	db := database.NewDynamoDB(nil) // Pass nil to let it create its own client when needed
	
	// Create handler with database
	h := handler.NewHandler(db)

	var response events.LambdaFunctionURLResponse
	var routeErr error

	switch {
	case method == http.MethodPost && path == "/shorten":
		response, routeErr = h.ShortenURL(ctx, event)
	
	case method == http.MethodGet && strings.HasPrefix(path, "/stats/"):
		response, routeErr = h.GetURLStats(ctx, event)
	
	case method == http.MethodGet && path != "/":
		// Any other GET request is treated as a redirect
		response, routeErr = h.RedirectURL(ctx, event)
	
	default:
		logger.Warn("Route not found", map[string]interface{}{
			"method": method,
			"path":   path,
		})
		response = events.LambdaFunctionURLResponse{
			StatusCode: http.StatusNotFound,
			Body:       `{"error": "Not found"}`,
		}
	}

	// Record overall API latency
	if metricClient != nil {
		latencyMs := float64(time.Since(startTime).Milliseconds())
		endpoint := "unknown"
		switch {
		case method == http.MethodPost && path == "/shorten":
			endpoint = "/shorten"
		case method == http.MethodGet && strings.HasPrefix(path, "/stats/"):
			endpoint = "/stats/{shortCode}"
		case method == http.MethodGet && path != "/":
			endpoint = "/{shortCode}"
		default:
			endpoint = path
		}
		
		metricClient.RecordAPILatency(ctx, endpoint, latencyMs)
	}

	return response, routeErr
}

func main() {
	logger.Info("URL Shortener Lambda starting up")
	lambda.Start(router)
}
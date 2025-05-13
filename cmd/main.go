package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/handler"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/logger"
)

func router(ctx context.Context, event events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	path := event.RawPath
	method := event.RequestContext.HTTP.Method

	logger.Info("Received request", map[string]interface{}{
		"method":    method,
		"path":      path,
		"requestId": event.RequestContext.RequestID,
		"source":    event.RequestContext.HTTP.SourceIP,
	})

	switch {
	case method == http.MethodPost && path == "/shorten":
		return handler.ShortenURL(ctx, event)
	
	case method == http.MethodGet && strings.HasPrefix(path, "/stats/"):
		return handler.GetURLStats(ctx, event)
	
	case method == http.MethodGet && path != "/":
		// Any other GET request is treated as a redirect
		return handler.RedirectURL(ctx, event)
	
	default:
		logger.Warn("Route not found", map[string]interface{}{
			"method": method,
			"path":   path,
		})
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusNotFound,
			Body:       `{"error": "Not found"}`,
		}, nil
	}
}

func main() {
	logger.Info("URL Shortener Lambda starting up")
	lambda.Start(router)
}
package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/handler"
)

func router(ctx context.Context, event events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	path := event.RawPath
	method := event.RequestContext.HTTP.Method

	switch {
	case method == http.MethodPost && path == "/shorten":
		return handler.ShortenURL(ctx, event)
	
	case method == http.MethodGet && strings.HasPrefix(path, "/stats/"):
		return handler.GetURLStats(ctx, event)
	
	case method == http.MethodGet && path != "/":
		// Any other GET request is treated as a redirect
		return handler.RedirectURL(ctx, event)
	
	default:
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusNotFound,
			Body:       `{"error": "Not found"}`,
		}, nil
	}
}

func main() {
	lambda.Start(router)
}
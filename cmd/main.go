package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"https://github.com/yusupscopes/aws-url-shortener-api/pkg/handler"
	"net/http"
)

func router(ctx context.Context, event events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	switch event.RequestContext.HTTP.Method {
	case http.MethodPost:
		if event.RawPath == "/shorten" {
			return handler.ShortenURL(ctx, event)
		}
	case http.MethodGet:
		return handler.RedirectURL(ctx, event)
	default:
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Body:       "Method Not Allowed",
		}, nil
	}

	return events.LambdaFunctionURLResponse{
		StatusCode: http.StatusNotFound,
		Body:       "Not Found",
	}, nil
}

func main() {
	lambda.Start(router)
}

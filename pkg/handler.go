package handler

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
)

func ShortenURL(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	return events.LambdaFunctionURLResponse{
		StatusCode: http.StatusCreated,
		Body:       `{"message":"Shorten URL placeholder"}`,
	}, nil
}

func RedirectURL(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	return events.LambdaFunctionURLResponse{
		StatusCode: http.StatusFound,
		Headers: map[string]string{
			"Location": "https://example.com",
		},
	}, nil
}

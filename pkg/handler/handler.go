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
)

const (
	// Length of the generated short code
	codeLength = 5
)

// ShortenURL handles the creation of a new short URL
func ShortenURL(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	// Initialize DynamoDB client
	client, err := database.GetClient(ctx)
	if err != nil {
		fmt.Printf("Error initializing DynamoDB client: %v\n", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"error": "Internal server error: %v"}`, err),
		}, nil
	}

	// Parse request body
	var shortenReq model.ShortenRequest
	err = json.Unmarshal([]byte(req.Body), &shortenReq)
	if err != nil {
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "Invalid request body"}`,
		}, nil
	}

	if shortenReq.URL == "" {
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "URL is required"}`,
		}, nil
	}

	// Generate a random code for the short URL
	code, err := utils.GenerateShortCode(codeLength)
	if err != nil {
		fmt.Printf("Error generating short code: %v\n", err)
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
		fmt.Printf("Error creating URL: %v\n", err)
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
	fmt.Printf("Successfully created short URL: %s for original URL: %s\n", shortURL, urlItem.OriginalURL)

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
	// Initialize DynamoDB client
	client, err := database.GetClient(ctx)
	if err != nil {
		fmt.Printf("Error initializing DynamoDB client: %v\n", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"error": "Internal server error: %v"}`, err),
		}, nil
	}

	// Extract code from path
	path := req.RawPath
	if path == "/" {
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "Short code is required"}`,
		}, nil
	}

	// Remove leading slash
	code := strings.TrimPrefix(path, "/")

	// Get URL from DynamoDB
	urlItem, err := database.GetURL(ctx, client, code)
	if err != nil {
		fmt.Printf("Error retrieving URL: %v\n", err)
		if strings.Contains(err.Error(), "URL not found") {
			return events.LambdaFunctionURLResponse{
				StatusCode: http.StatusNotFound,
				Body:       `{"error": "URL not found"}`,
			}, nil
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
			fmt.Printf("Error incrementing click count: %v\n", err)
		}
	}()

	fmt.Printf("Successfully retrieved original URL: %s for code: %s\n", urlItem.OriginalURL, code)

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
	// Initialize DynamoDB client
	client, err := database.GetClient(ctx)
	if err != nil {
		fmt.Printf("Error initializing DynamoDB client: %v\n", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"error": "Internal server error: %v"}`, err),
		}, nil
	}

	// Extract code from path
	path := req.RawPath
	code := strings.TrimPrefix(path, "/stats/")
	if code == "" || code == path {
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "Short code is required"}`,
		}, nil
	}

	// Get URL from DynamoDB
	urlItem, err := database.GetURL(ctx, client, code)
	if err != nil {
		fmt.Printf("Error retrieving URL: %v\n", err)
		if strings.Contains(err.Error(), "URL not found") {
			return events.LambdaFunctionURLResponse{
				StatusCode: http.StatusNotFound,
				Body:       `{"error": "URL not found"}`,
			}, nil
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

	responseJSON, _ := json.Marshal(stats)
	return events.LambdaFunctionURLResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJSON),
	}, nil
}
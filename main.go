package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const (
	// Length of the generated short code
	codeLength = 5
	// Characters used in the random short code
	codeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// Table name for DynamoDB
	tableName = "UrlShortener"
)

// URL represents the URL item in DynamoDB
type URL struct {
	ID  string `json:"id" dynamodbav:"id"`
	URL string `json:"url" dynamodbav:"url"`
}

// URLRequest represents the request body for creating a new short URL
type URLRequest struct {
	URL string `json:"url"`
}

// URLResponse represents the response for creating a new short URL
type URLResponse struct {
	OriginalURL string `json:"originalUrl"`
	ShortURL    string `json:"shortUrl"`
}

// generateRandomCode generates a random short code of specified length
func generateRandomCode(length int) (string, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	charsLength := len(codeChars)
	for i := 0; i < length; i++ {
		buffer[i] = codeChars[int(buffer[i])%charsLength]
	}

	return string(buffer), nil
}

// getClient creates and returns a DynamoDB client
func getClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg), nil
}

// createURL creates a new short URL
func createURL(ctx context.Context, client *dynamodb.Client, originalURL string) (*URL, error) {
	// Generate a random code for the short URL
	code, err := generateRandomCode(codeLength)
	if err != nil {
		return nil, err
	}

	urlItem := URL{
		ID:  code,
		URL: originalURL,
	}

	// Marshal URL item to DynamoDB attribute values
	av, err := attributevalue.MarshalMap(urlItem)
	if err != nil {
		return nil, err
	}

	// Put item into DynamoDB
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	})
	if err != nil {
		return nil, err
	}

	return &urlItem, nil
}

// getURL retrieves a URL by its short code
func getURL(ctx context.Context, client *dynamodb.Client, code string) (*URL, error) {
	// Get key condition
	key, err := attributevalue.MarshalMap(map[string]string{
		"id": code,
	})
	if err != nil {
		return nil, err
	}

	// Get item from DynamoDB
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	})
	if err != nil {
		return nil, err
	}

	if len(result.Item) == 0 {
		return nil, fmt.Errorf("URL not found for code: %s", code)
	}

	var urlItem URL
	err = attributevalue.UnmarshalMap(result.Item, &urlItem)
	if err != nil {
		return nil, err
	}

	return &urlItem, nil
}

// Handler is the Lambda function handler
func Handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize DynamoDB client
	client, err := getClient(ctx)
	if err != nil {
		log.Printf("Error initializing DynamoDB client: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"error": "Internal server error: %v"}`, err),
		}, nil
	}

	// Get base URL from environment variable or use a default
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		// Extract base URL from the request
		baseURL = fmt.Sprintf("https://%s/%s", event.RequestContext.DomainName, event.RequestContext.Stage)
	}

	// Route based on the HTTP method and path
	switch {
	case event.HTTPMethod == "POST" && event.Path == "/url":
		// Create a new short URL
		var urlRequest URLRequest
		err := json.Unmarshal([]byte(event.Body), &urlRequest)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       `{"error": "Invalid request body"}`,
			}, nil
		}

		if urlRequest.URL == "" {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       `{"error": "URL is required"}`,
			}, nil
		}

		urlItem, err := createURL(ctx, client, urlRequest.URL)
		if err != nil {
			log.Printf("Error creating URL: %v", err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       fmt.Sprintf(`{"error": "Failed to create short URL: %v"}`, err),
			}, nil
		}

		shortURL := fmt.Sprintf("%s/url/%s", baseURL, urlItem.ID)
		response := URLResponse{
			OriginalURL: urlItem.URL,
			ShortURL:    shortURL,
		}

		responseJSON, _ := json.Marshal(response)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: string(responseJSON),
		}, nil

	case event.HTTPMethod == "GET" && strings.HasPrefix(event.Path, "/url/"):
		// Get original URL from short code
		code := strings.TrimPrefix(event.Path, "/url/")
		if code == "" {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       `{"error": "Code is required"}`,
			}, nil
		}

		urlItem, err := getURL(ctx, client, code)
		if err != nil {
			log.Printf("Error retrieving URL: %v", err)
			if strings.Contains(err.Error(), "URL not found") {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusNotFound,
					Body:       fmt.Sprintf(`{"error": "%v"}`, err),
				}, nil
			}
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       fmt.Sprintf(`{"error": "Failed to retrieve URL: %v"}`, err),
			}, nil
		}

		// Redirect to the original URL
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusFound,
			Headers: map[string]string{
				"Location": urlItem.URL,
			},
			Body: "",
		}, nil

	default:
		// Handle unknown routes
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       `{"error": "Not found"}`,
		}, nil
	}
}

func main() {
	lambda.Start(Handler)
}
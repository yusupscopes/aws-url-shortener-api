package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/logger"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/model"
)

const (
	// Table name for DynamoDB
	TableName = "UrlShortener"
)

// GetClient creates and returns a DynamoDB client
func GetClient(ctx context.Context) (*dynamodb.Client, error) {
	logger.Debug("Initializing DynamoDB client")
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("Failed to load AWS config", err)
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg), nil
}

// CreateURL creates a new short URL in DynamoDB
func CreateURL(ctx context.Context, client *dynamodb.Client, urlItem *model.URLItem) error {
	logger.Debug("Creating URL in DynamoDB", map[string]interface{}{
		"shortCode": urlItem.ShortCode,
	})
	
	// Marshal URL item to DynamoDB attribute values
	av, err := attributevalue.MarshalMap(urlItem)
	if err != nil {
		logger.Error("Failed to marshal URL item", map[string]interface{}{
			"error": err.Error(),
			"item":  urlItem,
		})
		return err
	}

	// Put item into DynamoDB
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(TableName),
		Item:      av,
	})
	
	if err != nil {
		logger.Error("Failed to put item in DynamoDB", map[string]interface{}{
			"error":     err.Error(),
			"shortCode": urlItem.ShortCode,
			"tableName": TableName,
		})
		return err
	}
	
	logger.Debug("Successfully created URL in DynamoDB", map[string]interface{}{
		"shortCode": urlItem.ShortCode,
		"tableName": TableName,
	})
	return nil
}

// GetURL retrieves a URL by its short code
func GetURL(ctx context.Context, client *dynamodb.Client, code string) (*model.URLItem, error) {
	logger.Debug("Getting URL from DynamoDB", map[string]interface{}{
		"shortCode": code,
		"tableName": TableName,
	})
	
	// Get key condition
	key, err := attributevalue.MarshalMap(map[string]string{
		"shortCode": code,
	})
	if err != nil {
		logger.Error("Failed to marshal key", map[string]interface{}{
			"error":     err.Error(),
			"shortCode": code,
		})
		return nil, err
	}

	// Get item from DynamoDB
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key:       key,
	})
	if err != nil {
		logger.Error("Failed to get item from DynamoDB", map[string]interface{}{
			"error":     err.Error(),
			"shortCode": code,
			"tableName": TableName,
		})
		return nil, err
	}

	if len(result.Item) == 0 {
		logger.Warn("URL not found in DynamoDB", map[string]interface{}{
			"shortCode": code,
			"tableName": TableName,
		})
		return nil, fmt.Errorf("URL not found for code: %s", code)
	}

	var urlItem model.URLItem
	err = attributevalue.UnmarshalMap(result.Item, &urlItem)
	if err != nil {
		logger.Error("Failed to unmarshal DynamoDB item", map[string]interface{}{
			"error":     err.Error(),
			"shortCode": code,
			"item":      result.Item,
		})
		return nil, err
	}

	logger.Debug("Successfully retrieved URL from DynamoDB", map[string]interface{}{
		"shortCode": code,
		"tableName": TableName,
	})
	return &urlItem, nil
}

// IncrementClickCount increments the click count for a URL
func IncrementClickCount(ctx context.Context, client *dynamodb.Client, code string) error {
	logger.Debug("Incrementing click count in DynamoDB", map[string]interface{}{
		"shortCode": code,
		"tableName": TableName,
	})
	
	key, err := attributevalue.MarshalMap(map[string]string{
		"shortCode": code,
	})
	if err != nil {
		logger.Error("Failed to marshal key for click count update", map[string]interface{}{
			"error":     err.Error(),
			"shortCode": code,
		})
		return err
	}

	_, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(TableName),
		Key:       key,
		UpdateExpression: aws.String("SET clickCount = clickCount + :inc"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inc": &types.AttributeValueMemberN{Value: "1"},
		},
	})
	
	if err != nil {
		logger.Error("Failed to update click count in DynamoDB", map[string]interface{}{
			"error":     err.Error(),
			"shortCode": code,
			"tableName": TableName,
		})
		return err
	}
	
	logger.Debug("Successfully incremented click count in DynamoDB", map[string]interface{}{
		"shortCode": code,
		"tableName": TableName,
	})
	return nil
}
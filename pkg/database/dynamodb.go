package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/model"
)

const (
	// Table name for DynamoDB
	TableName = "UrlShortener"
)

// GetClient creates and returns a DynamoDB client
func GetClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg), nil
}

// CreateURL creates a new short URL in DynamoDB
func CreateURL(ctx context.Context, client *dynamodb.Client, urlItem *model.URLItem) error {
	// Marshal URL item to DynamoDB attribute values
	av, err := attributevalue.MarshalMap(urlItem)
	if err != nil {
		return err
	}

	// Put item into DynamoDB
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(TableName),
		Item:      av,
	})
	return err
}

// GetURL retrieves a URL by its short code
func GetURL(ctx context.Context, client *dynamodb.Client, code string) (*model.URLItem, error) {
	// Get key condition
	key, err := attributevalue.MarshalMap(map[string]string{
		"shortCode": code,
	})
	if err != nil {
		return nil, err
	}

	// Get item from DynamoDB
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key:       key,
	})
	if err != nil {
		return nil, err
	}

	if len(result.Item) == 0 {
		return nil, fmt.Errorf("URL not found for code: %s", code)
	}

	var urlItem model.URLItem
	err = attributevalue.UnmarshalMap(result.Item, &urlItem)
	if err != nil {
		return nil, err
	}

	return &urlItem, nil
}

// IncrementClickCount increments the click count for a URL
func IncrementClickCount(ctx context.Context, client *dynamodb.Client, code string) error {
	key, err := attributevalue.MarshalMap(map[string]string{
		"shortCode": code,
	})
	if err != nil {
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
	return err
}
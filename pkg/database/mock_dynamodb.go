package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/model"
)

// MockDynamoDB is a mock implementation of DynamoDB for testing
type MockDynamoDB struct {
	urls     map[string]*model.URLItem
	mutex    sync.RWMutex
	failNext bool
}

// NewMockDynamoDB creates a new mock DynamoDB client
func NewMockDynamoDB() DynamoDBInterface {
	return &MockDynamoDB{
		urls: make(map[string]*model.URLItem),
	}
}

// SetFailNext makes the next operation fail
func (m *MockDynamoDB) SetFailNext(fail bool) {
	m.failNext = fail
}

// GetClient returns a mock DynamoDB client
func (m *MockDynamoDB) GetClient(ctx context.Context) (*dynamodb.Client, error) {
	if m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("mock error: failed to get client")
	}
	
	// Return nil as we won't use the actual client in tests
	return nil, nil
}

// CreateURL mocks creating a URL in DynamoDB
func (m *MockDynamoDB) CreateURL(ctx context.Context, urlItem *model.URLItem) error {
	if m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: failed to create URL")
	}
	
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Store a copy of the URL item
	m.urls[urlItem.ShortCode] = &model.URLItem{
		ShortCode:   urlItem.ShortCode,
		OriginalURL: urlItem.OriginalURL,
		CreatedAt:   urlItem.CreatedAt,
		Expiration:  urlItem.Expiration,
		ClickCount:  urlItem.ClickCount,
	}
	
	return nil
}

// GetURL mocks retrieving a URL from DynamoDB
func (m *MockDynamoDB) GetURL(ctx context.Context, code string) (*model.URLItem, error) {
	if m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("mock error: failed to get URL")
	}
	
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	urlItem, exists := m.urls[code]
	if !exists {
		return nil, fmt.Errorf("URL not found for code: %s", code)
	}
	
	// Return a copy of the URL item
	return &model.URLItem{
		ShortCode:   urlItem.ShortCode,
		OriginalURL: urlItem.OriginalURL,
		CreatedAt:   urlItem.CreatedAt,
		Expiration:  urlItem.Expiration,
		ClickCount:  urlItem.ClickCount,
	}, nil
}

// IncrementClickCount mocks incrementing the click count
func (m *MockDynamoDB) IncrementClickCount(ctx context.Context, code string) error {
	if m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: failed to increment click count")
	}
	
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	urlItem, exists := m.urls[code]
	if !exists {
		return fmt.Errorf("URL not found for code: %s", code)
	}
	
	urlItem.ClickCount++
	return nil
}
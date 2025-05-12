package model

// URLItem represents the URL item in DynamoDB
type URLItem struct {
	ShortCode   string `json:"shortCode" dynamodbav:"shortCode"`
	OriginalURL string `json:"originalURL" dynamodbav:"originalURL"`
	CreatedAt   string `json:"createdAt" dynamodbav:"createdAt"`
	Expiration  int64  `json:"expiration,omitempty" dynamodbav:"expiration,omitempty"`
	ClickCount  int    `json:"clickCount" dynamodbav:"clickCount"`
}

// ShortenRequest represents the request body for creating a new short URL
type ShortenRequest struct {
	URL          string `json:"url"`
	ExpireInDays int    `json:"expire_in_days,omitempty"`
}

// ShortenResponse represents the response for creating a new short URL
type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

// StatsResponse represents the analytics response for a short URL
type StatsResponse struct {
	OriginalURL string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
	Expiration  int64  `json:"expiration,omitempty"`
	ClickCount  int    `json:"click_count"`
}
package model

type ShortURL struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
	Expiration  int64  `json:"expiration,omitempty"`
	ClickCount  int64  `json:"click_count"`
}

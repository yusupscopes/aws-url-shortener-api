package utils

import (
	"crypto/rand"
	"time"
)

const (
	// Characters used in the random short code
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// GenerateShortCode generates a random short code of specified length
func GenerateShortCode(length int) (string, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	charsLength := len(charset)
	for i := 0; i < length; i++ {
		buffer[i] = charset[int(buffer[i])%charsLength]
	}

	return string(buffer), nil
}

// CalculateExpirationTime calculates the expiration timestamp based on days
func CalculateExpirationTime(days int) int64 {
	if days <= 0 {
		return 0 // No expiration
	}
	return time.Now().AddDate(0, 0, days).Unix()
}
package utils

import (
	"testing"
	"time"
)

func TestGenerateShortCode(t *testing.T) {
	// Test code length
	length := 5
	code, err := GenerateShortCode(length)
	if err != nil {
		t.Fatalf("GenerateShortCode returned an error: %v", err)
	}
	
	if len(code) != length {
		t.Errorf("Expected code length %d, got %d", length, len(code))
	}
	
	// Test uniqueness (basic check)
	code2, err := GenerateShortCode(length)
	if err != nil {
		t.Fatalf("GenerateShortCode returned an error: %v", err)
	}
	
	if code == code2 {
		t.Errorf("Expected unique codes, got the same code twice: %s", code)
	}
	
	// Test character set
	for _, char := range code {
		if !contains(charset, char) {
			t.Errorf("Code contains invalid character: %c", char)
		}
	}
}

func contains(s string, c rune) bool {
	for _, char := range s {
		if char == c {
			return true
		}
	}
	return false
}

func TestCalculateExpirationTime(t *testing.T) {
	// Test with zero days (no expiration)
	expiration := CalculateExpirationTime(0)
	if expiration != 0 {
		t.Errorf("Expected 0 for no expiration, got %d", expiration)
	}
	
	// Test with negative days (no expiration)
	expiration = CalculateExpirationTime(-1)
	if expiration != 0 {
		t.Errorf("Expected 0 for negative days, got %d", expiration)
	}
	
	// Test with positive days
	days := 7
	expiration = CalculateExpirationTime(days)
	
	// Calculate expected expiration (approximately)
	expectedExpiration := time.Now().AddDate(0, 0, days).Unix()
	
	// Allow for a small difference due to execution time
	diff := expiration - expectedExpiration
	if diff < -5 || diff > 5 {
		t.Errorf("Expected expiration around %d, got %d (diff: %d)", expectedExpiration, expiration, diff)
	}
}
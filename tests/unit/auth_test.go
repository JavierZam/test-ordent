package unit

import (
	"testing"
	"time"

	"test-ordent/internal/auth"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test_secret"
	userID := uint(1)
	role := "admin"
	expiry := time.Hour * 24

	// Generate token
	token, err := auth.GenerateToken(userID, role, secret, expiry)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate token
	claims, err := auth.ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Check claims
	if claims.UserID != userID {
		t.Errorf("Expected UserID to be %d, got %d", userID, claims.UserID)
	}

	if claims.Role != role {
		t.Errorf("Expected Role to be %s, got %s", role, claims.Role)
	}

	// Check invalid token
	_, err = auth.ValidateToken("invalid.token.string", secret)
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}

	// Check wrong secret
	_, err = auth.ValidateToken(token, "wrong_secret")
	if err == nil {
		t.Error("Expected error for wrong secret, got nil")
	}
}
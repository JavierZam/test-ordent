package unit

import (
	"testing"
	"time"

	"test-ordent/internal/auth"
)

func TestGenerateToken(t *testing.T) {
	testCases := []struct {
		name     string
		userID   uint
		role     string
		secret   string
		expected bool
	}{
		{
			name:     "Valid token generation",
			userID:   1,
			role:     "admin",
			secret:   "test_secret",
			expected: true,
		},
		{
			name:     "Empty secret",
			userID:   1,
			role:     "admin",
			secret:   "",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := auth.GenerateToken(tc.userID, tc.role, tc.secret, 24*time.Hour)
			
			if tc.expected && (err != nil || token == "") {
				t.Errorf("Expected token generation to succeed, but got error: %v", err)
			}
			
			if !tc.expected && (err == nil && token != "") {
				t.Errorf("Expected token generation to fail, but it succeeded")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	secret := "test_secret"
	userID := uint(1)
	role := "admin"
	
	validToken, err := auth.GenerateToken(userID, role, secret, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token for testing: %v", err)
	}
	
	expiredToken, err := auth.GenerateToken(userID, role, secret, -1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate expired token for testing: %v", err)
	}
	
	testCases := []struct {
		name     string
		token    string
		secret   string
		expected bool
	}{
		{
			name:     "Valid token",
			token:    validToken,
			secret:   secret,
			expected: true,
		},
		{
			name:     "Invalid token format",
			token:    "invalid.token.format",
			secret:   secret,
			expected: false,
		},
		{
			name:     "Empty token",
			token:    "",
			secret:   secret,
			expected: false,
		},
		{
			name:     "Wrong secret",
			token:    validToken,
			secret:   "wrong_secret",
			expected: false,
		},
		{
			name:     "Expired token",
			token:    expiredToken,
			secret:   secret,
			expected: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := auth.ValidateToken(tc.token, tc.secret)
			
			if tc.expected {
				if err != nil {
					t.Errorf("Expected token validation to succeed, but got error: %v", err)
				}
				
				if claims == nil {
					t.Errorf("Expected claims to be non-nil")
				} else {
					if claims.UserID != userID {
						t.Errorf("Expected UserID to be %d, got %d", userID, claims.UserID)
					}
					
					if claims.Role != role {
						t.Errorf("Expected Role to be %s, got %s", role, claims.Role)
					}
				}
			} else {
				if err == nil {
					t.Errorf("Expected token validation to fail, but it succeeded")
				}
			}
		})
	}
}
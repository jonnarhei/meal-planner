package auth

import (
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-secret"
	var userID int64 = 1 
	email := "test@test.com"
	expiry := 86400

	token, err := GenerateToken(userID, email, secret, expiry)
	if err != nil {
		t.Fatalf("expected no error generating token, got %v", err)
	}

	if token == "" {
		t.Fatal("exptected token not to be empty")
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("exptected no error validating token, got %v", err)
	}

	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}
}

func TestValidateTokenWithWrongSecret(t *testing.T) {
	token, err := GenerateToken(1, "test@test.com", "correct-secret", 86400)
	if err != nil {
		t.Fatalf("expected no error generating token, got %v", err)
	}

	_, err = ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error validating token with wrong secret, received nil")
	}
}

func TestValidateExpiredToken(t *testing.T) {
	token, err := GenerateToken(1, "test@test.com", "secret", -1)
	if err != nil {
		t.Fatalf("expected no error generating token, got %v", err)
	}

	time.Sleep(5 * time.Second)

	_, err = ValidateToken(token, "secret")
	if err == nil {
		t.Fatal("exptected error validating expired token, received nil")
	}
}
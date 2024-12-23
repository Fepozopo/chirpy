package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "mySecretPassword"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if len(hashedPassword) == 0 {
		t.Fatal("Hashed password is empty")
	}

	err = CheckPasswordHash(password, hashedPassword)
	if err != nil {
		t.Fatalf("Failed to check password hash: %v", err)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "mySecretPassword"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	err = CheckPasswordHash(password, hashedPassword)
	if err != nil {
		t.Fatalf("Failed to check password hash: %v", err)
	}

	err = CheckPasswordHash("wrongPassword", hashedPassword)
	if err == nil {
		t.Fatal("CheckPasswordHash should have returned an error")
	}
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mySecret"

	token, err := MakeJWT(userID, tokenSecret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	if len(token) == 0 {
		t.Fatal("Generated JWT is empty")
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mySecret"

	token, err := MakeJWT(userID, tokenSecret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	validUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if validUserID != userID {
		t.Fatal("Validated user ID is not the same as the generated one")
	}

	invalidToken := "Invalid Token"
	_, err = ValidateJWT(invalidToken, tokenSecret)
	if err == nil {
		t.Fatal("ValidateJWT should have returned an error")
	}
}

func TestGetBearerToken(t *testing.T) {
	validToken := "myValidToken"
	headers := http.Header{
		"Authorization": []string{"Bearer " + validToken},
	}

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("Failed to retrieve bearer token: %v", err)
	}

	if token != validToken {
		t.Fatal("Retrieved token is not the same as the valid one")
	}

	invalidHeaders := http.Header{}
	_, err = GetBearerToken(invalidHeaders)
	if err == nil {
		t.Fatal("GetBearerToken should have returned an error")
	}

	invalidHeaders = http.Header{
		"Invalid": []string{"Bearer " + validToken},
	}

	_, err = GetBearerToken(invalidHeaders)
	if err == nil {
		t.Fatal("GetBearerToken should have returned an error")
	}
}

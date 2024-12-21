package auth

import (
	"testing"
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

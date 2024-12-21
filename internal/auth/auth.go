package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword takes a password string and returns a hashed string of the
// password and an error. The hashed string is a string of bytes generated
// by the bcrypt.GenerateFromPassword() function. The error is returned if
// there is an error generating the hash.
func HashPassword(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hashBytes), nil
}

// CheckPasswordHash checks if the given password matches the given hash.
// It returns an error if the password does not match the hash.
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

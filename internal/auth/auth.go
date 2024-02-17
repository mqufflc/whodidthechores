package auth

import (
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return "", nil
	}
	return string(hashedPassword), nil
}

func VerifyPassword(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(password), []byte(hash))
}

func checkPasswordComplexity(password string) error {
	const minEntropyBits = 65
	return passwordvalidator.Validate(password, minEntropyBits)
}

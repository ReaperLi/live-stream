package util

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword returns the bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password:%w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword checks if provided password is correct or not
func CheckPassword(password string, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return errors.New("密码错误")
	}
	return nil
}

func IsPasswordConfirmed(password, passwordConfirm string) bool {
	if password != "" && password == passwordConfirm {
		return true
	}
	return false
}

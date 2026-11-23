package auth

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

// bcryptCost — 12 итераций: около 250 мс на хеш на обычном железе.
const bcryptCost = 12

// HashPassword возвращает bcrypt-хеш пароля.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", apperr.Internal(err)
	}
	return string(hash), nil
}

// CheckPassword сравнивает пароль с хешем за постоянное время.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ValidatePassword проверяет политику пароля: 8–72 байта (ограничение bcrypt).
func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return apperr.Validation("пароль должен содержать не менее 8 символов")
	}
	if len(password) > 72 {
		return apperr.Validation("пароль длиннее 72 байт не поддерживается")
	}
	if strings.TrimSpace(password) == "" {
		return apperr.Validation("пароль не может состоять из пробелов")
	}
	return nil
}

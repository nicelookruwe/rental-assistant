// Package tenants — арендаторы арендодателя.
package tenants

import (
	"net/mail"
	"strings"
	"time"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

// Tenant — арендатор пользователя.
type Tenant struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Input — данные создания и обновления арендатора.
type Input struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

// Normalize обрезает пробелы.
func (in *Input) Normalize() {
	in.Name = strings.TrimSpace(in.Name)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Email = strings.TrimSpace(in.Email)
}

// Validate проверяет обязательные поля и форматы.
func (in Input) Validate() error {
	switch {
	case in.Name == "":
		return apperr.Validation("укажите имя арендатора")
	case len([]rune(in.Name)) > 200:
		return apperr.Validation("имя длиннее 200 символов")
	case len([]rune(in.Phone)) > 40:
		return apperr.Validation("телефон длиннее 40 символов")
	}
	if in.Email != "" {
		addr, err := mail.ParseAddress(in.Email)
		if err != nil || !strings.EqualFold(addr.Address, in.Email) {
			return apperr.Validation("укажите корректный email арендатора")
		}
	}
	return nil
}

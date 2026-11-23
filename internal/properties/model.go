// Package properties — квартиры арендодателя.
package properties

import (
	"strings"
	"time"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

// Property — объект недвижимости, принадлежащий пользователю.
type Property struct {
	ID          string    `json:"id"`
	UserID      string    `json:"-"`
	Address     string    `json:"address"`
	Area        float64   `json:"area"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Input — данные создания и полного обновления квартиры.
type Input struct {
	Address     string  `json:"address"`
	Area        float64 `json:"area"`
	Description string  `json:"description"`
}

// Normalize обрезает пробелы в строковых полях.
func (in *Input) Normalize() {
	in.Address = strings.TrimSpace(in.Address)
	in.Description = strings.TrimSpace(in.Description)
}

// Validate проверяет бизнес-ограничения квартиры.
func (in Input) Validate() error {
	switch {
	case in.Address == "":
		return apperr.Validation("укажите адрес квартиры")
	case len([]rune(in.Address)) > 500:
		return apperr.Validation("адрес длиннее 500 символов")
	case in.Area <= 0:
		return apperr.Validation("площадь должна быть больше нуля")
	case in.Area > 10000:
		return apperr.Validation("площадь выглядит нереалистично (больше 10 000 м²)")
	case len([]rune(in.Description)) > 2000:
		return apperr.Validation("описание длиннее 2000 символов")
	}
	return nil
}

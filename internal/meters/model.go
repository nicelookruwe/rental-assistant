// Package meters — история показаний счётчиков.
package meters

import (
	"strings"
	"time"

	"github.com/ailina/rental-assistant/pkg/apperr"
	"github.com/ailina/rental-assistant/pkg/ids"
)

// Типы счётчиков.
const (
	TypeElectricity = "electricity"
	TypeWaterCold   = "cold_water"
	TypeWaterHot    = "hot_water"
	TypeGas         = "gas"
	TypeHeating     = "heating"
	TypeOther       = "other"
)

// DateLayout — формат дат в API.
const DateLayout = "2006-01-02"

// Reading — показание счётчика. Строки только добавляются.
type Reading struct {
	ID              string    `json:"id"`
	PropertyID      string    `json:"property_id"`
	PropertyAddress string    `json:"property_address"`
	MeterType       string    `json:"meter_type"`
	Value           float64   `json:"value"`
	RecordedAt      string    `json:"recorded_at"`
	Note            string    `json:"note"`
	CreatedAt       time.Time `json:"created_at"`
}

// Latest — последнее показание по каждому счётчику квартиры.
type Latest struct {
	PropertyID      string  `json:"property_id"`
	PropertyAddress string  `json:"property_address"`
	MeterType       string  `json:"meter_type"`
	Value           float64 `json:"value"`
	RecordedAt      string  `json:"recorded_at"`
}

// Input — новое показание.
type Input struct {
	PropertyID string  `json:"property_id"`
	MeterType  string  `json:"meter_type"`
	Value      float64 `json:"value"`
	RecordedAt string  `json:"recorded_at"`
	Note       string  `json:"note"`
}

// Normalize чистит строковые поля.
func (in *Input) Normalize() {
	in.PropertyID = strings.TrimSpace(in.PropertyID)
	in.MeterType = strings.ToLower(strings.TrimSpace(in.MeterType))
	in.RecordedAt = strings.TrimSpace(in.RecordedAt)
	in.Note = strings.TrimSpace(in.Note)
}

// Validate проверяет квартиру, тип, значение и дату.
func (in Input) Validate() error {
	if !ids.Valid(in.PropertyID) {
		return apperr.Validation("выберите квартиру для показания")
	}
	switch in.MeterType {
	case TypeElectricity, TypeWaterCold, TypeWaterHot, TypeGas, TypeHeating, TypeOther:
	default:
		return apperr.Validation("неизвестный тип счётчика: %s", in.MeterType)
	}
	if in.Value < 0 {
		return apperr.Validation("показание не может быть отрицательным")
	}
	if in.Value > 1_000_000 {
		return apperr.Validation("показание выглядит нереалистично")
	}
	recorded, err := time.Parse(DateLayout, in.RecordedAt)
	if err != nil {
		return apperr.Validation("дата показания должна быть в формате ГГГГ-ММ-ДД")
	}
	if recorded.After(time.Now().AddDate(0, 0, 1)) {
		return apperr.Validation("дата показания не может быть в будущем")
	}
	if len([]rune(in.Note)) > 500 {
		return apperr.Validation("примечание длиннее 500 символов")
	}
	return nil
}

package contracts

import "testing"

const (
	testProperty = "0b7f2c1e-8a51-4c2b-9f3d-2f7a6c1d4e10"
	testTenant   = "5f3e9a7c-1b2d-4e6f-8a90-7c3b1d2e4f56"
)

func validInput() Input {
	return Input{
		PropertyID: testProperty,
		TenantID:   testTenant,
		StartDate:  "2026-01-01",
		EndDate:    "2026-12-31",
		RentAmount: 45000,
		Deposit:    45000,
	}
}

func TestNormalizeDefaultsToActive(t *testing.T) {
	in := validInput()
	in.Status = "  "
	in.Normalize()
	if in.Status != StatusActive {
		t.Fatalf("статус по умолчанию = %q, ожидался %q", in.Status, StatusActive)
	}
}

func TestValidateAcceptsCorrectContract(t *testing.T) {
	in := validInput()
	in.Normalize()
	if err := in.Validate(); err != nil {
		t.Fatalf("корректный договор отклонён: %v", err)
	}
}

func TestValidateRejectsBrokenData(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Input)
	}{
		{"пустая квартира", func(in *Input) { in.PropertyID = "" }},
		{"пустой арендатор", func(in *Input) { in.TenantID = "" }},
		{"формат даты", func(in *Input) { in.StartDate = "01.01.2026" }},
		{"конец раньше начала", func(in *Input) { in.EndDate = "2025-12-31" }},
		{"конец равен началу", func(in *Input) { in.EndDate = in.StartDate }},
		{"отрицательная аренда", func(in *Input) { in.RentAmount = -1 }},
		{"отрицательный залог", func(in *Input) { in.Deposit = -5 }},
		{"неизвестный статус", func(in *Input) { in.Status = "podpisan" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validInput()
			tc.mutate(&in)
			if err := in.Validate(); err == nil {
				t.Fatal("ожидалась ошибка валидации")
			}
		})
	}
}

func TestValidateRejectsTooLongContract(t *testing.T) {
	in := validInput()
	in.EndDate = "2090-01-01"
	if err := in.Validate(); err == nil {
		t.Fatal("ожидалась ошибка: срок больше 50 лет")
	}
}

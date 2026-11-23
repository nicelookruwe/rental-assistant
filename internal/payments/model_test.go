package payments

import "testing"

func TestValidatePaymentInput(t *testing.T) {
	base := Input{
		ContractID:  testContract,
		Amount:      45000,
		PaymentDate: "2026-11-05",
		Type:        TypeRent,
		Status:      StatusPaid,
	}

	cases := []struct {
		name   string
		mutate func(*Input)
		ok     bool
	}{
		{"корректный", func(in *Input) {}, true},
		{"пустой договор", func(in *Input) { in.ContractID = "" }, false},
		{"нулевая сумма", func(in *Input) { in.Amount = 0 }, false},
		{"отрицательная сумма", func(in *Input) { in.Amount = -100 }, false},
		{"слишком большая сумма", func(in *Input) { in.Amount = 500_000_000 }, false},
		{"плохая дата", func(in *Input) { in.PaymentDate = "05/11/2026" }, false},
		{"неизвестный тип", func(in *Input) { in.Type = "ipoteka" }, false},
		{"просрочка вручную", func(in *Input) { in.Status = StatusOverdue }, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := base
			tc.mutate(&in)
			err := in.Validate()
			if tc.ok && err != nil {
				t.Fatalf("ожидался успех, получена ошибка: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("ожидалась ошибка валидации")
			}
		})
	}
}

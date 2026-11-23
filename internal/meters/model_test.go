package meters

import "testing"

const testPropertyID = "0b7f2c1e-8a51-4c2b-9f3d-2f7a6c1d4e10"

func TestValidateAcceptsReading(t *testing.T) {
	in := Input{PropertyID: testPropertyID, MeterType: TypeElectricity, Value: 15234.5, RecordedAt: "2026-11-01"}
	in.Normalize()
	if err := in.Validate(); err != nil {
		t.Fatalf("корректное показание отклонено: %v", err)
	}
}

func TestNormalizeLowercasesMeterType(t *testing.T) {
	in := Input{PropertyID: testPropertyID, MeterType: " Cold_Water ", Value: 12, RecordedAt: "2026-11-01"}
	in.Normalize()
	if in.MeterType != TypeWaterCold {
		t.Fatalf("тип счётчика = %q, ожидался %q", in.MeterType, TypeWaterCold)
	}
}

func TestValidateRejectsBrokenReadings(t *testing.T) {
	cases := []struct {
		name string
		in   Input
	}{
		{"нет квартиры", Input{MeterType: TypeGas, Value: 1, RecordedAt: "2026-11-01"}},
		{"неизвестный тип", Input{PropertyID: testPropertyID, MeterType: "svet", Value: 1, RecordedAt: "2026-11-01"}},
		{"отрицательное значение", Input{PropertyID: testPropertyID, MeterType: TypeGas, Value: -1, RecordedAt: "2026-11-01"}},
		{"слишком большое", Input{PropertyID: testPropertyID, MeterType: TypeGas, Value: 2_000_000, RecordedAt: "2026-11-01"}},
		{"будущая дата", Input{PropertyID: testPropertyID, MeterType: TypeGas, Value: 1, RecordedAt: "2099-01-01"}},
		{"плохая дата", Input{PropertyID: testPropertyID, MeterType: TypeGas, Value: 1, RecordedAt: "01.11.2026"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := tc.in
			in.Normalize()
			if err := in.Validate(); err == nil {
				t.Fatal("ожидалась ошибка валидации")
			}
		})
	}
}

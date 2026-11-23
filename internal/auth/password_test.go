package auth

import "testing"

func TestHashPasswordIsNotPlaintext(t *testing.T) {
	hash, err := HashPassword("peredacha-2026")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "peredacha-2026" {
		t.Fatal("пароль сохранён в открытом виде")
	}
	if !CheckPassword(hash, "peredacha-2026") {
		t.Fatal("верный пароль не прошёл проверку")
	}
	if CheckPassword(hash, "peredacha-2027") {
		t.Fatal("неверный пароль прошёл проверку")
	}
}

func TestHashPasswordIsSalted(t *testing.T) {
	first, err := HashPassword("povtor-12345")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	second, err := HashPassword("povtor-12345")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if first == second {
		t.Fatal("одинаковые пароли дали одинаковый хеш: соль не используется")
	}
}

func TestValidatePassword(t *testing.T) {
	cases := map[string]bool{
		"":               false,
		"корот":          false,
		"        ":       false,
		"normalnyi-2026": true,
	}
	for password, wantOK := range cases {
		err := ValidatePassword(password)
		if wantOK && err != nil {
			t.Errorf("пароль %q: неожиданная ошибка %v", password, err)
		}
		if !wantOK && err == nil {
			t.Errorf("пароль %q: ожидалась ошибка валидации", password)
		}
	}
}

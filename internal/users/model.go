package users

import "time"

// User — владелец недвижимости и единственный тип субъекта в системе.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

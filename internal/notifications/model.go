// Package notifications — напоминания о платежах и сроках договоров.
package notifications

import "time"

// Статусы уведомления.
const (
	StatusPending = "pending"
	StatusRead    = "read"
)

// Notification — напоминание пользователю.
type Notification struct {
	ID          string    `json:"id"`
	Message     string    `json:"message"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

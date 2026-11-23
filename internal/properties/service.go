package properties

import (
	"context"
)

// Repository-интерфейс нужен сервису, чтобы бизнес-логику можно было тестировать
// без PostgreSQL.
type store interface {
	List(ctx context.Context, userID string) ([]Property, error)
	Get(ctx context.Context, userID, id string) (Property, error)
	Create(ctx context.Context, userID string, in Input) (Property, error)
	Update(ctx context.Context, userID, id string, in Input) (Property, error)
	Delete(ctx context.Context, userID, id string) error
}

// Service — бизнес-логика квартир.
type Service struct {
	repo store
}

// NewService собирает сервис квартир.
func NewService(repo store) *Service { return &Service{repo: repo} }

// List возвращает квартиры пользователя.
func (s *Service) List(ctx context.Context, userID string) ([]Property, error) {
	return s.repo.List(ctx, userID)
}

// Get возвращает одну квартиру.
func (s *Service) Get(ctx context.Context, userID, id string) (Property, error) {
	return s.repo.Get(ctx, userID, id)
}

// Create проверяет данные и создаёт квартиру.
func (s *Service) Create(ctx context.Context, userID string, in Input) (Property, error) {
	in.Normalize()
	if err := in.Validate(); err != nil {
		return Property{}, err
	}
	return s.repo.Create(ctx, userID, in)
}

// Update проверяет данные и обновляет квартиру.
func (s *Service) Update(ctx context.Context, userID, id string, in Input) (Property, error) {
	in.Normalize()
	if err := in.Validate(); err != nil {
		return Property{}, err
	}
	return s.repo.Update(ctx, userID, id, in)
}

// Delete удаляет квартиру пользователя.
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	return s.repo.Delete(ctx, userID, id)
}

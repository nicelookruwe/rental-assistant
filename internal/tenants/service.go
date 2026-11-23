package tenants

import "context"

type store interface {
	List(ctx context.Context, userID string) ([]Tenant, error)
	Get(ctx context.Context, userID, id string) (Tenant, error)
	Create(ctx context.Context, userID string, in Input) (Tenant, error)
	Update(ctx context.Context, userID, id string, in Input) (Tenant, error)
	Delete(ctx context.Context, userID, id string) error
}

// Service — бизнес-логика арендаторов.
type Service struct {
	repo store
}

// NewService собирает сервис арендаторов.
func NewService(repo store) *Service { return &Service{repo: repo} }

// List возвращает арендаторов пользователя.
func (s *Service) List(ctx context.Context, userID string) ([]Tenant, error) {
	return s.repo.List(ctx, userID)
}

// Get возвращает одного арендатора.
func (s *Service) Get(ctx context.Context, userID, id string) (Tenant, error) {
	return s.repo.Get(ctx, userID, id)
}

// Create проверяет данные и создаёт арендатора.
func (s *Service) Create(ctx context.Context, userID string, in Input) (Tenant, error) {
	in.Normalize()
	if err := in.Validate(); err != nil {
		return Tenant{}, err
	}
	return s.repo.Create(ctx, userID, in)
}

// Update проверяет данные и обновляет арендатора.
func (s *Service) Update(ctx context.Context, userID, id string, in Input) (Tenant, error) {
	in.Normalize()
	if err := in.Validate(); err != nil {
		return Tenant{}, err
	}
	return s.repo.Update(ctx, userID, id, in)
}

// Delete удаляет арендатора.
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	return s.repo.Delete(ctx, userID, id)
}

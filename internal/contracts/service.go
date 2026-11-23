package contracts

import "context"

type store interface {
	List(ctx context.Context, userID string, f Filter) ([]Contract, error)
	Get(ctx context.Context, userID, id string) (Contract, error)
	Create(ctx context.Context, userID string, in Input) (Contract, error)
	Update(ctx context.Context, userID, id string, in Input) (Contract, error)
	Delete(ctx context.Context, userID, id string) error
	EndingSoon(ctx context.Context, userID string, days int) ([]Contract, error)
}

// Service — бизнес-логика договоров.
type Service struct {
	repo store
}

// NewService собирает сервис договоров.
func NewService(repo store) *Service { return &Service{repo: repo} }

// List возвращает договоры с фильтрами.
func (s *Service) List(ctx context.Context, userID string, f Filter) ([]Contract, error) {
	return s.repo.List(ctx, userID, f)
}

// Get возвращает договор.
func (s *Service) Get(ctx context.Context, userID, id string) (Contract, error) {
	return s.repo.Get(ctx, userID, id)
}

// Create проверяет данные и создаёт договор.
func (s *Service) Create(ctx context.Context, userID string, in Input) (Contract, error) {
	in.Normalize()
	if err := in.Validate(); err != nil {
		return Contract{}, err
	}
	return s.repo.Create(ctx, userID, in)
}

// Update проверяет данные и обновляет договор.
func (s *Service) Update(ctx context.Context, userID, id string, in Input) (Contract, error) {
	in.Normalize()
	if err := in.Validate(); err != nil {
		return Contract{}, err
	}
	return s.repo.Update(ctx, userID, id, in)
}

// Delete удаляет договор.
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	return s.repo.Delete(ctx, userID, id)
}

// EndingSoon возвращает договоры, истекающие в ближайшие days дней.
func (s *Service) EndingSoon(ctx context.Context, userID string, days int) ([]Contract, error) {
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	return s.repo.EndingSoon(ctx, userID, days)
}

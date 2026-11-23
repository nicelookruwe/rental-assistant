package meters

import "context"

type store interface {
	List(ctx context.Context, userID string, f Filter) ([]Reading, error)
	Get(ctx context.Context, userID, id string) (Reading, error)
	Latest(ctx context.Context, userID, propertyID string) ([]Latest, error)
	Create(ctx context.Context, userID string, in Input) (Reading, error)
}

// Service — бизнес-логика показаний счётчиков.
type Service struct {
	repo store
}

// NewService собирает сервис показаний.
func NewService(repo store) *Service { return &Service{repo: repo} }

// List возвращает историю показаний.
func (s *Service) List(ctx context.Context, userID string, f Filter) ([]Reading, error) {
	return s.repo.List(ctx, userID, f)
}

// Get возвращает показание.
func (s *Service) Get(ctx context.Context, userID, id string) (Reading, error) {
	return s.repo.Get(ctx, userID, id)
}

// Latest возвращает последние показания по счётчикам.
func (s *Service) Latest(ctx context.Context, userID, propertyID string) ([]Latest, error) {
	return s.repo.Latest(ctx, userID, propertyID)
}

// Create проверяет данные и добавляет показание (история не перезаписывается).
func (s *Service) Create(ctx context.Context, userID string, in Input) (Reading, error) {
	in.Normalize()
	if err := in.Validate(); err != nil {
		return Reading{}, err
	}
	return s.repo.Create(ctx, userID, in)
}

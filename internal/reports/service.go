package reports

import (
	"context"

	"github.com/ailina/rental-assistant/internal/contracts"
	"github.com/ailina/rental-assistant/internal/payments"
)

// Ограничения выборок для дашборда.
const (
	upcomingDays   = 30
	upcomingLimit  = 10
	overdueLimit   = 10
	recentLimit    = 10
	monthlyMonths  = 6
	endingSoonDays = 30
)

type store interface {
	Counts(ctx context.Context, userID string) (Counters, error)
	MonthMoney(ctx context.Context, userID string) (Money, error)
	OverdueTotal(ctx context.Context, userID string) (float64, int, error)
	Upcoming(ctx context.Context, userID string, days, limit int) ([]payments.Payment, error)
	Overdue(ctx context.Context, userID string, limit int) ([]payments.Payment, error)
	Recent(ctx context.Context, userID string, limit int) ([]payments.Payment, error)
	Monthly(ctx context.Context, userID string, months int) ([]MonthPoint, error)
}

type contractSource interface {
	EndingSoon(ctx context.Context, userID string, days int) ([]contracts.Contract, error)
}

// Service собирает сводку для главного экрана.
type Service struct {
	repo      store
	contracts contractSource
}

// NewService создаёт сервис отчётов.
func NewService(repo store, con contractSource) *Service {
	return &Service{repo: repo, contracts: con}
}

// Summary формирует полную сводку пользователя.
func (s *Service) Summary(ctx context.Context, userID string) (Summary, error) {
	var out Summary

	counts, err := s.repo.Counts(ctx, userID)
	if err != nil {
		return out, err
	}
	out.Counters = counts

	money, err := s.repo.MonthMoney(ctx, userID)
	if err != nil {
		return out, err
	}
	out.Money = money

	overdueAmount, overdueCount, err := s.repo.OverdueTotal(ctx, userID)
	if err != nil {
		return out, err
	}
	out.OverdueAmount = overdueAmount
	out.OverdueCount = overdueCount

	if out.Upcoming, err = s.repo.Upcoming(ctx, userID, upcomingDays, upcomingLimit); err != nil {
		return out, err
	}
	if out.Overdue, err = s.repo.Overdue(ctx, userID, overdueLimit); err != nil {
		return out, err
	}
	if out.Recent, err = s.repo.Recent(ctx, userID, recentLimit); err != nil {
		return out, err
	}
	if out.Monthly, err = s.repo.Monthly(ctx, userID, monthlyMonths); err != nil {
		return out, err
	}
	if out.EndingSoon, err = s.contracts.EndingSoon(ctx, userID, endingSoonDays); err != nil {
		return out, err
	}
	out.ContractsEndingSoon = len(out.EndingSoon)

	return out, nil
}

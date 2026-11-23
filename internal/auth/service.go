// Package auth — регистрация, вход и проверка JWT.
package auth

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/ailina/rental-assistant/internal/users"
	"github.com/ailina/rental-assistant/pkg/apperr"
)

// userStore — часть репозитория пользователей, нужная сервису аутентификации.
type userStore interface {
	Create(ctx context.Context, email, passwordHash string) (users.User, error)
	ByEmail(ctx context.Context, email string) (users.User, error)
	ByID(ctx context.Context, id string) (users.User, error)
}

// Result — результат успешной регистрации или входа.
type Result struct {
	User      users.User `json:"user"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
}

// Service реализует бизнес-логику аутентификации.
type Service struct {
	users  userStore
	tokens *TokenManager
}

// NewService собирает сервис аутентификации.
func NewService(users userStore, tokens *TokenManager) *Service {
	return &Service{users: users, tokens: tokens}
}

// Register создаёт пользователя и сразу выдаёт токен.
func (s *Service) Register(ctx context.Context, email, password string) (Result, error) {
	email = strings.TrimSpace(email)
	addr, err := mail.ParseAddress(email)
	if err != nil || !strings.EqualFold(addr.Address, email) {
		return Result{}, apperr.Validation("укажите корректный email")
	}
	if err := ValidatePassword(password); err != nil {
		return Result{}, err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return Result{}, err
	}
	user, err := s.users.Create(ctx, email, hash)
	if err != nil {
		return Result{}, err
	}
	return s.issue(user)
}

// Login проверяет пароль и выдаёт токен. Одинаковый ответ на неизвестную почту
// и неверный пароль не даёт перебирать зарегистрированные адреса.
func (s *Service) Login(ctx context.Context, email, password string) (Result, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return Result{}, apperr.Validation("укажите email и пароль")
	}
	user, err := s.users.ByEmail(ctx, email)
	if err != nil {
		if e, ok := apperr.As(err); ok && e.Kind == apperr.KindNotFound {
			return Result{}, apperr.Unauthorized("неверный email или пароль")
		}
		return Result{}, err
	}
	if !CheckPassword(user.PasswordHash, password) {
		return Result{}, apperr.Unauthorized("неверный email или пароль")
	}
	return s.issue(user)
}

// Profile возвращает текущего пользователя.
func (s *Service) Profile(ctx context.Context, userID string) (users.User, error) {
	return s.users.ByID(ctx, userID)
}

func (s *Service) issue(user users.User) (Result, error) {
	token, exp, err := s.tokens.Issue(user.ID)
	if err != nil {
		return Result{}, err
	}
	return Result{User: user, Token: token, ExpiresAt: exp}, nil
}

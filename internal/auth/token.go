package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

// Claims — полезная нагрузка токена (JWT, HS256).
type Claims struct {
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Issuer    string `json:"iss"`
}

const tokenIssuer = "rental-assistant"

// TokenManager выпускает и проверяет JWT.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenManager требует секрет не короче 16 байт.
func NewTokenManager(secret string, ttl time.Duration) (*TokenManager, error) {
	if len(secret) < 16 {
		return nil, errors.New("auth: JWT_SECRET должен быть не короче 16 символов")
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &TokenManager{secret: []byte(secret), ttl: ttl}, nil
}

// Issue подписывает токен для пользователя.
func (m *TokenManager) Issue(userID string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(m.ttl)
	claims := Claims{Subject: userID, IssuedAt: now.Unix(), ExpiresAt: exp.Unix(), Issuer: tokenIssuer}

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, apperr.Internal(err)
	}
	signingInput := header + "." + base64.RawURLEncoding.EncodeToString(payload)
	signature := m.sign(signingInput)
	return signingInput + "." + signature, exp, nil
}

// Parse проверяет подпись, срок действия и издателя; возвращает ID пользователя.
func (m *TokenManager) Parse(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", apperr.Unauthorized("некорректный токен")
	}
	signingInput := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(m.sign(signingInput))) {
		return "", apperr.Unauthorized("подпись токена неверна")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", apperr.Unauthorized("некорректный токен")
	}
	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return "", apperr.Unauthorized("некорректный токен")
	}
	if claims.Issuer != tokenIssuer {
		return "", apperr.Unauthorized("некорректный токен")
	}
	if time.Now().Unix() >= claims.ExpiresAt {
		return "", apperr.Unauthorized("срок действия токена истёк")
	}
	if claims.Subject == "" {
		return "", apperr.Unauthorized("некорректный токен")
	}
	return claims.Subject, nil
}

func (m *TokenManager) sign(input string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// String помогает в логах не печатать секрет.
func (m *TokenManager) String() string { return fmt.Sprintf("TokenManager(ttl=%s)", m.ttl) }

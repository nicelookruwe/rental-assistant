// Package ids — проверка и генерация UUID.
package ids

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Valid сообщает, похожа ли строка на UUID.
func Valid(s string) bool { return uuidRe.MatchString(strings.ToLower(strings.TrimSpace(s))) }

// New генерирует случайный UUID v4.
func New() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32], nil
}

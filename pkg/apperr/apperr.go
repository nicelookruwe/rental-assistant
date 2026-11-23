// Package apperr описывает предметные ошибки приложения и их категории.
// HTTP-слой преобразует категорию ошибки в код ответа, поэтому бизнес-логика
// не зависит от net/http.
package apperr

import (
	"errors"
	"fmt"
)

// Kind — категория ошибки.
type Kind int

const (
	KindInternal Kind = iota
	KindValidation
	KindUnauthorized
	KindForbidden
	KindNotFound
	KindConflict
)

// Error — ошибка приложения с категорией и человекочитаемым сообщением.
type Error struct {
	Kind    Kind
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

// Validation — неверные входные данные.
func Validation(format string, args ...any) *Error {
	return &Error{Kind: KindValidation, Message: fmt.Sprintf(format, args...)}
}

// Unauthorized — отсутствует или недействителен токен/учётные данные.
func Unauthorized(msg string) *Error {
	return &Error{Kind: KindUnauthorized, Message: msg}
}

// Forbidden — объект существует, но принадлежит другому пользователю.
func Forbidden() *Error {
	return &Error{Kind: KindForbidden, Message: "доступ запрещён"}
}

// NotFound — объект не найден (в том числе чужие объекты, чтобы не раскрывать их наличие).
func NotFound(entity string) *Error {
	return &Error{Kind: KindNotFound, Message: entity + " не найден"}
}

// Conflict — нарушение уникальности или состояния.
func Conflict(format string, args ...any) *Error {
	return &Error{Kind: KindConflict, Message: fmt.Sprintf(format, args...)}
}

// Internal — непредвиденная ошибка; детали остаются в логе.
func Internal(err error) *Error {
	return &Error{Kind: KindInternal, Message: "внутренняя ошибка сервера", Err: err}
}

// As достаёт *Error из цепочки.
func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

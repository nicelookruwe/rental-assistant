// Package storage — абстракция файлового хранилища.
//
// В MVP используется локальный диск (Local). Интерфейс Storage позволяет позже
// подключить MinIO/S3 без изменения модуля documents.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Storage сохраняет файлы по ключу, отдаёт их наружу и умеет удалять.
type Storage interface {
	// Save записывает содержимое и возвращает URL, по которому файл доступен.
	Save(ctx context.Context, key string, r io.Reader) (url string, err error)
	// Open открывает ранее сохранённый файл для чтения.
	Open(ctx context.Context, url string) (io.ReadSeekCloser, error)
	Delete(ctx context.Context, key string) error
}

// Local хранит файлы в каталоге root и отдаёт их по префиксу publicPrefix.
type Local struct {
	root         string
	publicPrefix string
}

// NewLocal создаёт локальное хранилище; каталог создаётся при необходимости.
func NewLocal(root, publicPrefix string) (*Local, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("storage: пустой каталог")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, fmt.Errorf("storage: создание каталога: %w", err)
	}
	if publicPrefix == "" {
		publicPrefix = "/files"
	}
	return &Local{root: abs, publicPrefix: strings.TrimSuffix(publicPrefix, "/")}, nil
}

// Root возвращает абсолютный путь каталога (нужен для раздачи файлов).
func (l *Local) Root() string { return l.root }

// Save пишет файл. Ключ чистится от «..», чтобы нельзя было выйти за root.
func (l *Local) Save(_ context.Context, key string, r io.Reader) (string, error) {
	clean, err := cleanKey(key)
	if err != nil {
		return "", err
	}
	full := filepath.Join(l.root, filepath.FromSlash(clean))
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		return "", fmt.Errorf("storage: каталог файла: %w", err)
	}
	f, err := os.OpenFile(full, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o640)
	if err != nil {
		return "", fmt.Errorf("storage: запись файла: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("storage: копирование: %w", err)
	}
	return l.publicPrefix + "/" + clean, nil
}

// Open открывает файл по URL, ранее выданному Save.
func (l *Local) Open(_ context.Context, url string) (io.ReadSeekCloser, error) {
	clean, err := cleanKey(strings.TrimPrefix(url, l.publicPrefix+"/"))
	if err != nil {
		return nil, err
	}
	f, err := os.Open(filepath.Join(l.root, filepath.FromSlash(clean)))
	if err != nil {
		return nil, fmt.Errorf("storage: открытие файла: %w", err)
	}
	return f, nil
}

// Delete удаляет файл по URL, ранее выданному Save.
func (l *Local) Delete(_ context.Context, url string) error {
	clean, err := cleanKey(strings.TrimPrefix(url, l.publicPrefix+"/"))
	if err != nil {
		return err
	}
	err = os.Remove(filepath.Join(l.root, filepath.FromSlash(clean)))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func cleanKey(key string) (string, error) {
	key = strings.TrimPrefix(key, "/")
	if key == "" || strings.Contains(key, "..") {
		return "", errors.New("storage: некорректный ключ файла")
	}
	return key, nil
}

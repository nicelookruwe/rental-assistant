package documents

import (
	"context"
	"io"
	"strings"

	"github.com/ailina/rental-assistant/pkg/apperr"
	"github.com/ailina/rental-assistant/pkg/ids"
	"github.com/ailina/rental-assistant/pkg/storage"
)

type store interface {
	List(ctx context.Context, userID, propertyID string) ([]Document, error)
	Get(ctx context.Context, userID, id string) (Document, error)
	Create(ctx context.Context, userID string, in Input) (Document, error)
	Delete(ctx context.Context, userID, id string) error
}

// Service — бизнес-логика документов: сохраняет файл и метаданные.
type Service struct {
	repo  store
	files storage.Storage
}

// NewService собирает сервис документов.
func NewService(repo store, files storage.Storage) *Service {
	return &Service{repo: repo, files: files}
}

// List возвращает документы пользователя.
func (s *Service) List(ctx context.Context, userID, propertyID string) ([]Document, error) {
	if propertyID != "" && !ids.Valid(propertyID) {
		return nil, apperr.Validation("некорректный property_id")
	}
	return s.repo.List(ctx, userID, propertyID)
}

// Get возвращает метаданные документа.
func (s *Service) Get(ctx context.Context, userID, id string) (Document, error) {
	return s.repo.Get(ctx, userID, id)
}

// Upload сохраняет файл и создаёт запись. Если запись не создалась, файл удаляется.
func (s *Service) Upload(ctx context.Context, userID, propertyID, docType, fileName string, content io.Reader, size int64) (Document, error) {
	in := Input{PropertyID: propertyID, DocumentType: docType, FileName: fileName, SizeBytes: size}
	in.Normalize()
	if err := in.Validate(); err != nil {
		return Document{}, err
	}

	key, err := ids.New()
	if err != nil {
		return Document{}, apperr.Internal(err)
	}
	objectKey := userID + "/" + key + "_" + SafeFileName(in.FileName)

	url, err := s.files.Save(ctx, objectKey, content)
	if err != nil {
		return Document{}, apperr.Internal(err)
	}
	in.FileURL = url

	doc, err := s.repo.Create(ctx, userID, in)
	if err != nil {
		// метаданные не сохранились — не оставляем мусор в хранилище
		_ = s.files.Delete(ctx, url)
		return Document{}, err
	}
	return doc, nil
}

// Open отдаёт содержимое файла документа, проверяя владельца.
func (s *Service) Open(ctx context.Context, userID, id string) (Document, io.ReadSeekCloser, error) {
	doc, err := s.repo.Get(ctx, userID, id)
	if err != nil {
		return Document{}, nil, err
	}
	f, err := s.files.Open(ctx, doc.FileURL)
	if err != nil {
		return Document{}, nil, apperr.NotFound("файл документа")
	}
	return doc, f, nil
}

// Delete удаляет запись и файл.
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	doc, err := s.repo.Get(ctx, userID, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, userID, id); err != nil {
		return err
	}
	if strings.TrimSpace(doc.FileURL) != "" {
		if err := s.files.Delete(ctx, doc.FileURL); err != nil {
			return apperr.Internal(err)
		}
	}
	return nil
}

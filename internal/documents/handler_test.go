package documents

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ailina/rental-assistant/internal/auth"
	"github.com/ailina/rental-assistant/pkg/apperr"
)

const testDocumentID = "8b486a11-0dbd-4ac0-96b7-34a374721892"

type handlerStore struct {
	doc    Document
	userID string
}

func (s *handlerStore) List(context.Context, string, string) ([]Document, error) { return nil, nil }
func (s *handlerStore) Get(_ context.Context, userID, id string) (Document, error) {
	if userID != s.userID || id != s.doc.ID {
		return Document{}, apperr.NotFound("документ")
	}
	return s.doc, nil
}
func (s *handlerStore) Create(context.Context, string, Input) (Document, error) {
	return Document{}, nil
}
func (s *handlerStore) Delete(context.Context, string, string) error { return nil }

type memoryStorage struct {
	data      []byte
	openCalls int
}

func (s *memoryStorage) Save(context.Context, string, io.Reader) (string, error) { return "", nil }
func (s *memoryStorage) Open(context.Context, string) (io.ReadSeekCloser, error) {
	s.openCalls++
	return readSeekCloser{Reader: bytes.NewReader(s.data)}, nil
}
func (s *memoryStorage) Delete(context.Context, string) error { return nil }

type readSeekCloser struct{ *bytes.Reader }

func (readSeekCloser) Close() error { return nil }

func downloadHandler(t *testing.T, owner string, files *memoryStorage) (http.Handler, string, *auth.TokenManager) {
	t.Helper()
	tokens, err := auth.NewTokenManager("test-secret-with-at-least-16-bytes", time.Hour)
	if err != nil {
		t.Fatalf("NewTokenManager: %v", err)
	}
	token, _, err := tokens.Issue(owner)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	store := &handlerStore{userID: owner, doc: Document{
		ID: testDocumentID, FileName: "contract.pdf", FileURL: "/doc", CreatedAt: time.Now(),
	}}
	handler := auth.Middleware(tokens)(http.HandlerFunc(NewHandler(NewService(store, files)).Download))
	return handler, token, tokens
}

func TestDownloadForcesSafeAttachmentHeaders(t *testing.T) {
	files := &memoryStorage{data: []byte("%PDF-1.7\ncontent")}
	handler, token, _ := downloadHandler(t, "owner", files)
	req := httptest.NewRequest(http.MethodGet, "/api/documents/"+testDocumentID+"/file", nil)
	req.SetPathValue("id", testDocumentID)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body: %s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Disposition"); !strings.HasPrefix(got, "attachment;") {
		t.Fatalf("Content-Disposition = %q, expected attachment", got)
	}
	if got := res.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
	if got := res.Header().Get("Content-Type"); got != "application/octet-stream" {
		t.Fatalf("Content-Type = %q", got)
	}
}

func TestDownloadRejectsInvalidTokenBeforeStorageAccess(t *testing.T) {
	files := &memoryStorage{data: []byte("%PDF-1.7\ncontent")}
	handler, _, _ := downloadHandler(t, "owner", files)
	req := httptest.NewRequest(http.MethodGet, "/api/documents/"+testDocumentID+"/file", nil)
	req.SetPathValue("id", testDocumentID)
	req.Header.Set("Authorization", "Bearer invalid")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.Code)
	}
	if files.openCalls != 0 {
		t.Fatal("файл был открыт до проверки авторизации")
	}
}

func TestDownloadDoesNotOpenAnotherUsersDocument(t *testing.T) {
	files := &memoryStorage{data: []byte("%PDF-1.7\ncontent")}
	handler, _, tokens := downloadHandler(t, "owner", files)
	token, _, err := tokens.Issue("other-user")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/documents/"+testDocumentID+"/file", nil)
	req.SetPathValue("id", testDocumentID)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.Code)
	}
	if files.openCalls != 0 {
		t.Fatal("файл чужого пользователя был открыт")
	}
}

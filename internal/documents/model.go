// Package documents — файлы квартир: договоры, акты, чеки, фотографии.
package documents

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ailina/rental-assistant/pkg/apperr"
	"github.com/ailina/rental-assistant/pkg/ids"
)

// Типы документов.
const (
	TypeContract = "contract"
	TypeAct      = "act"
	TypeReceipt  = "receipt"
	TypePhoto    = "photo"
	TypeOther    = "other"
)

// MaxUploadSize — предел размера файла (20 МиБ).
const MaxUploadSize = 20 << 20

const sniffSize = 512

// Document — метаданные файла, привязанного к квартире.
type Document struct {
	ID              string    `json:"id"`
	PropertyID      string    `json:"property_id"`
	PropertyAddress string    `json:"property_address"`
	FileURL         string    `json:"file_url"`
	FileName        string    `json:"file_name"`
	DocumentType    string    `json:"document_type"`
	SizeBytes       int64     `json:"size_bytes"`
	CreatedAt       time.Time `json:"created_at"`
}

// Input — сохраняемый документ.
type Input struct {
	PropertyID   string
	DocumentType string
	FileName     string
	FileURL      string
	SizeBytes    int64
}

// Normalize чистит строковые поля и подставляет тип по умолчанию.
func (in *Input) Normalize() {
	in.PropertyID = strings.TrimSpace(in.PropertyID)
	in.DocumentType = strings.ToLower(strings.TrimSpace(in.DocumentType))
	in.FileName = strings.TrimSpace(in.FileName)
	if in.DocumentType == "" {
		in.DocumentType = TypeOther
	}
}

// Validate проверяет тип, квартиру и имя файла.
func (in Input) Validate() error {
	if !ids.Valid(in.PropertyID) {
		return apperr.Validation("выберите квартиру для документа")
	}
	switch in.DocumentType {
	case TypeContract, TypeAct, TypeReceipt, TypePhoto, TypeOther:
	default:
		return apperr.Validation("неизвестный тип документа: %s", in.DocumentType)
	}
	if in.FileName == "" {
		return apperr.Validation("не указано имя файла")
	}
	if len([]rune(in.FileName)) > 255 {
		return apperr.Validation("имя файла длиннее 255 символов")
	}
	if in.SizeBytes <= 0 {
		return apperr.Validation("файл пустой")
	}
	return nil
}

// ValidateFile проверяет имя, определённый MIME-тип и сигнатуру файла. В MVP
// поддерживаются форматы, которые нужны для договоров и фотографий, но не
// могут быть выполнены браузером как HTML/SVG/JavaScript.
func ValidateFile(name string, content io.ReadSeeker) (string, error) {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return "", apperr.Validation("у файла должно быть расширение")
	}

	buf := make([]byte, sniffSize)
	n, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return "", apperr.Validation("не удалось прочитать файл")
	}
	if _, err := content.Seek(0, io.SeekStart); err != nil {
		return "", apperr.Validation("не удалось проверить файл")
	}
	buf = buf[:n]
	mimeType := http.DetectContentType(buf)

	valid := false
	switch {
	case ext == ".pdf" && mimeType == "application/pdf" && bytes.HasPrefix(buf, []byte("%PDF-")):
		valid = true
	case (ext == ".jpg" || ext == ".jpeg") && mimeType == "image/jpeg" &&
		len(buf) >= 3 && bytes.Equal(buf[:3], []byte{0xff, 0xd8, 0xff}):
		valid = true
	case ext == ".png" && mimeType == "image/png" &&
		bytes.HasPrefix(buf, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		valid = true
	case ext == ".webp" && mimeType == "image/webp" &&
		len(buf) >= 12 && bytes.Equal(buf[:4], []byte("RIFF")) && bytes.Equal(buf[8:12], []byte("WEBP")):
		valid = true
	}
	if !valid {
		return "", apperr.Validation("поддерживаются только PDF, JPEG, PNG и WebP с корректным содержимым")
	}
	return mimeType, nil
}

// SafeFileName оставляет от имени файла только безопасные символы.
func SafeFileName(name string) string {
	base := filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '.', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('_')
		default:
			if r > 127 { // кириллица и другие буквы сохраняются
				b.WriteRune(r)
			}
		}
	}
	out := strings.Trim(b.String(), "._")
	if out == "" {
		out = "file"
	}
	if len([]rune(out)) > 80 {
		out = string([]rune(out)[:80])
	}
	return out
}

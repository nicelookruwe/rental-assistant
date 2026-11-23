package documents

import (
	"bytes"
	"testing"
)

func TestValidateFileAcceptsMatchingSafeFormats(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		mime string
	}{
		{"contract.pdf", []byte("%PDF-1.7\ncontent"), "application/pdf"},
		{"photo.jpg", []byte{0xff, 0xd8, 0xff, 0xe0, 0x00}, "image/jpeg"},
		{"photo.png", []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, "image/png"},
		{"photo.webp", []byte("RIFF\x10\x00\x00\x00WEBPVP8 "), "image/webp"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateFile(tt.name, bytes.NewReader(tt.data))
			if err != nil {
				t.Fatalf("ValidateFile: %v", err)
			}
			if got != tt.mime {
				t.Fatalf("MIME = %q, want %q", got, tt.mime)
			}
		})
	}
}

func TestValidateFileRejectsExecutableOrMismatchedContent(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"payload.html", []byte("<!doctype html><script>alert(1)</script>")},
		{"payload.svg", []byte(`<svg onload="alert(1)"/>`)},
		{"photo.jpg", []byte("<!doctype html><script>alert(1)</script>")},
		{"contract.pdf", []byte("not a PDF")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ValidateFile(tt.name, bytes.NewReader(tt.data)); err == nil {
				t.Fatal("небезопасный или несоответствующий файл принят")
			}
		})
	}
}

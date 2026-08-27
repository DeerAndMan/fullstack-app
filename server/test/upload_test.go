package test

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fullstack-app/server/pkg/upload"
)

func fileHeader(t *testing.T, name, contentType string, data []byte) *multipart.FileHeader {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatal(err)
	}
	fh := req.MultipartForm.File["file"][0]
	fh.Header.Set("Content-Type", contentType)
	return fh
}

func TestUploaderSave(t *testing.T) {
	base := t.TempDir()
	uploader := upload.NewUploader(base, 1, []string{".txt"})
	info, err := uploader.Save(fileHeader(t, "note.txt", "text/plain", []byte("hello uploader")))
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if info.FileName != "note.txt" || info.FileSize != int64(len("hello uploader")) || info.MimeType != "text/plain; charset=utf-8" {
		t.Fatalf("unexpected file info: %+v", info)
	}
	if filepath.IsAbs(info.FilePath) || !strings.HasSuffix(info.FilePath, ".txt") {
		t.Fatalf("FilePath should be a relative dated path: %q", info.FilePath)
	}
	stored := filepath.Join(base, info.FilePath)
	data, err := os.ReadFile(stored)
	if err != nil || string(data) != "hello uploader" {
		t.Fatalf("stored data = %q, err = %v", data, err)
	}
}

func TestUploaderRejectsTooLargeAndUnsupportedExtension(t *testing.T) {
	uploader := upload.NewUploader(t.TempDir(), 1, []string{".txt"})
	_, err := uploader.Save(fileHeader(t, "large.txt", "text/plain", bytes.Repeat([]byte("x"), 2*1024*1024)))
	if !errors.Is(err, upload.ErrFileTooLarge) {
		t.Errorf("large file error = %v, want ErrFileTooLarge", err)
	}

	_, err = uploader.Save(fileHeader(t, "image.jpg", "image/jpeg", []byte("not really an image")))
	if !errors.Is(err, upload.ErrFileTypeNotAllowed) {
		t.Errorf("extension error = %v, want ErrFileTypeNotAllowed", err)
	}
}

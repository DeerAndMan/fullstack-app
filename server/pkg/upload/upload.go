package upload

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Uploader struct {
	basePath  string
	maxSize   int64 // bytes
	allowExts []string
}

func NewUploader(basePath string, maxSizeMB int64, allowExts []string) *Uploader {
	return &Uploader{
		basePath:  basePath,
		maxSize:   maxSizeMB * 1024 * 1024,
		allowExts: allowExts,
	}
}

type FileInfo struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
}

func (u *Uploader) Save(header *multipart.FileHeader) (*FileInfo, error) {
	if header.Size > u.maxSize {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d bytes)", header.Size, u.maxSize)
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !slices.Contains(u.allowExts, ext) {
		return nil, fmt.Errorf("file type not allowed: %s", ext)
	}

	// date-based directory
	dateDir := time.Now().Format("2006/01/02")
	dir := filepath.Join(u.basePath, dateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	newName := uuid.New().String() + ext
	dstPath := filepath.Join(dir, newName)

	src, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	return &FileInfo{
		FileName: header.Filename,
		FilePath: filepath.Join(dateDir, newName),
		FileSize: header.Size,
		MimeType: header.Header.Get("Content-Type"),
	}, nil
}

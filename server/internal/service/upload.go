package service

import (
	"errors"
	"mime/multipart"

	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/upload"
)

type UploadService struct {
	uploader *upload.Uploader
}

func NewUploadService(uploader *upload.Uploader) *UploadService {
	return &UploadService{uploader: uploader}
}

func (s *UploadService) Upload(header *multipart.FileHeader) (*upload.FileInfo, error) {
	info, err := s.uploader.Save(header)
	if err != nil {
		if errors.Is(err, upload.ErrFileTooLarge) {
			return nil, errcode.ErrFileTooLarge
		}
		if errors.Is(err, upload.ErrFileTypeNotAllowed) {
			return nil, errcode.ErrFileTypeNotAllowed
		}
		return nil, errcode.ErrInternal
	}
	return info, nil
}

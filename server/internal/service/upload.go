package service

import (
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/upload"

	"mime/multipart"
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
		if err.Error() == "file too large" {
			return nil, errcode.ErrFileTooLarge
		}
		if err.Error() == "file type not allowed" {
			return nil, errcode.ErrFileTypeNotAllowed
		}
		return nil, errcode.ErrInternal
	}
	return info, nil
}

package service

import (
	"context"
	"fmt"
	"slices"

	"github.com/Falokut/storage_service/domain"
	"github.com/Falokut/storage_service/entity"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

//go:generate mockgen -source=repository.go -destination=mocks/imageStorage.go
type FileStorage interface {
	UploadFile(ctx context.Context, file entity.File) error
	GetFile(ctx context.Context, filename string, category string) (*entity.File, error)
	IsFileExist(ctx context.Context, filename string, category string) (bool, error)
	DeleteFile(ctx context.Context, filename string, category string) error
}

type Files struct {
	storage            FileStorage
	maxFileSize        int64
	supportedFileTypes []string
}

func NewFiles(storage FileStorage, maxFileSize int64, supportedFileTypes []string) Files {
	return Files{
		storage:            storage,
		maxFileSize:        maxFileSize,
		supportedFileTypes: supportedFileTypes,
	}
}

func (s Files) UploadFile(ctx context.Context, req domain.UploadFileRequest) (string, error) {
	fileSize := int64(len(req.Content))
	if fileSize == 0 {
		return "",
			domain.NewInvalidArgumentError("file has zero size", domain.ErrCodeFileHasZeroSize)
	}
	if fileSize > s.maxFileSize {
		return "",
			domain.NewInvalidArgumentError(
				fmt.Sprintf("file is too large max file size: %d, file size: %d",
					s.maxFileSize, fileSize),
				domain.ErrCodeFileTooBig,
			)
	}

	contentType := mimetype.Detect(req.Content).String()
	if len(s.supportedFileTypes) != 0 && !slices.Contains(s.supportedFileTypes, contentType) {
		return "", domain.NewInvalidArgumentError(
			fmt.Sprintf("file type is not supported. file type: '%s'", contentType),
			domain.ErrCodeUnsupportedFileType,
		)
	}

	filename := req.Filename
	if len(filename) == 0 {
		filename = uuid.NewString()
	}

	saveFileReq := entity.File{
		Filename:    filename,
		Category:    req.Category,
		ContentType: contentType,
		Content:     req.Content,
		Size:        int64(len(req.Content)),
	}
	err := s.storage.UploadFile(ctx, saveFileReq)
	if err != nil {
		return "", errors.WithMessage(err, "save file")
	}

	return filename, nil
}

func (s Files) GetFile(ctx context.Context, req domain.FileRequest) (*entity.File, error) {
	file, err := s.storage.GetFile(ctx, req.Filename, req.Category)
	if err != nil {
		return nil, errors.WithMessage(err, "get file")
	}
	return file, nil
}

func (s Files) IsFileExist(ctx context.Context, req domain.FileRequest) (bool, error) {
	exists, err := s.storage.IsFileExist(ctx, req.Filename, req.Category)
	if err != nil {
		return false, errors.WithMessage(err, "is file exist")
	}
	return exists, nil
}

func (s Files) DeleteFile(ctx context.Context, req domain.FileRequest) error {
	err := s.storage.DeleteFile(ctx, req.Filename, req.Category)
	if err != nil {
		return errors.WithMessage(err, "delete file")
	}
	return nil
}

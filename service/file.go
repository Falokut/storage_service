package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"slices"

	"storage-service/domain"
	"storage-service/entity"

	"github.com/Falokut/go-kit/http/types"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

//go:generate mockgen -source=repository.go -destination=mocks/imageStorage.go
type FileStorage interface {
	UploadFile(ctx context.Context, file entity.Metadata, reader io.Reader) error
	GetFile(ctx context.Context, filename string, category string, opt *types.RangeOption) (*entity.Metadata, io.ReadSeekCloser, error)
	IsFileExist(ctx context.Context, filename string, category string) (bool, error)
	DeleteFile(ctx context.Context, filename string, category string) error
}

type Pending interface {
	Enqueue(ctx context.Context, fileName string, category string) error
	Rollback(ctx context.Context, fileName string, category string) error
	Commit(ctx context.Context, fileName string, category string) error
}

type Files struct {
	storage            FileStorage
	supportedFileTypes []string
	pendingSrv         Pending
}

func NewFiles(storage FileStorage, supportedFileTypes []string, pendingSrv Pending) Files {
	return Files{
		storage:            storage,
		supportedFileTypes: supportedFileTypes,
		pendingSrv:         pendingSrv,
	}
}

func (s Files) UploadFile(ctx context.Context, req entity.UploadFileRequest) (string, error) {
	if req.ContentReader == nil {
		return "", domain.NewInvalidArgumentError("file has zero size", domain.ErrCodeFileHasZeroSize)
	}

	header := make([]byte, 512)
	n, _ := io.ReadFull(req.ContentReader, header)
	reader := io.MultiReader(bytes.NewReader(header[:n]), req.ContentReader)

	contentType := mimetype.Detect(header[:n]).String()

	if len(s.supportedFileTypes) != 0 && !slices.Contains(s.supportedFileTypes, contentType) {
		return "", domain.NewInvalidArgumentError(
			fmt.Sprintf("file type is not supported. file type: '%s'", contentType),
			domain.ErrCodeUnsupportedFileType,
		)
	}

	filename := req.Filename
	if filename == "" {
		filename = uuid.NewString()
	}

	metadata := entity.Metadata{
		Filename:    filename,
		PrettyName:  req.PrettyName,
		Category:    req.Category,
		ContentType: contentType,
		Size:        -1, // размер неизвестен заранее
	}

	// Если файл Pending
	if req.Pending {
		err := s.pendingSrv.Enqueue(ctx, filename, req.Category)
		if err != nil {
			return "", errors.WithMessage(err, "enqueue pending file")
		}
	}

	// Streaming upload в хранилище
	err := s.storage.UploadFile(ctx, metadata, reader)
	if err != nil {
		return "", errors.WithMessage(err, "save file")
	}

	return filename, nil
}

func (s Files) GetFile(
	ctx context.Context,
	req domain.FileRequest,
	opt *types.RangeOption,
) (*entity.Metadata, io.ReadSeekCloser, error) {
	metadata, contentReader, err := s.storage.GetFile(ctx, req.Filename, req.Category, opt)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "get file")
	}
	return metadata, contentReader, nil
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

func (s Files) Rollback(ctx context.Context, req domain.FileRequest) error {
	err := s.pendingSrv.Rollback(ctx, req.Filename, req.Category)
	if err != nil {
		return errors.WithMessage(err, "rollback file")
	}
	return nil
}

func (s Files) Commit(ctx context.Context, req domain.FileRequest) error {
	err := s.pendingSrv.Commit(ctx, req.Filename, req.Category)
	if err != nil {
		return errors.WithMessage(err, "commit file")
	}
	return nil
}
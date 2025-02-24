package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/Falokut/go-kit/http/types"
	"github.com/Falokut/storage_service/domain"
	"github.com/Falokut/storage_service/entity"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

//go:generate mockgen -source=repository.go -destination=mocks/imageStorage.go
type FileStorage interface {
	UploadFile(ctx context.Context, file entity.Metadata, reader io.Reader) error
	GetFile(ctx context.Context, filename string, category string, opt *types.RangeOption) (*entity.Metadata, io.Reader, error)
	IsFileExist(ctx context.Context, filename string, category string) (bool, error)
	DeleteFile(ctx context.Context, filename string, category string) error
}

type Files struct {
	storage               FileStorage
	maxRangeRequestLength int64
	maxFileSize           int64
	supportedFileTypes    []string
}

func NewFiles(storage FileStorage, maxFileSize int64, maxRangeRequestLength int64, supportedFileTypes []string) Files {
	return Files{
		storage:               storage,
		maxFileSize:           maxFileSize,
		maxRangeRequestLength: maxRangeRequestLength,
		supportedFileTypes:    supportedFileTypes,
	}
}

func (s Files) UploadFile(ctx context.Context, req entity.UploadFileRequest) (string, error) {
	if req.ContentReader == nil || req.Size <= 0 {
		return "",
			domain.NewInvalidArgumentError("file has zero size", domain.ErrCodeFileHasZeroSize)
	}
	if s.maxFileSize > 0 && req.Size >= s.maxFileSize {
		return "",
			domain.NewInvalidArgumentError("file size too big", domain.ErrCodeFileTooBig)
	}

	var readSeeker io.ReadSeeker
	if seeker, ok := req.ContentReader.(io.ReadSeeker); ok {
		readSeeker = seeker
	} else {
		buf, err := io.ReadAll(req.ContentReader)
		if err != nil {
			return "", errors.WithMessage(err, "read content for mime detection")
		}
		readSeeker = bytes.NewReader(buf)
	}
	contentTypeMime, err := mimetype.DetectReader(readSeeker)
	if err != nil {
		return "", errors.WithMessage(err, "detect file mime type")
	}
	_, err = readSeeker.Seek(0, io.SeekStart)
	if err != nil {
		return "", errors.WithMessage(err, "move reader to seek start")
	}

	contentType := contentTypeMime.String()
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
		Category:    req.Category,
		ContentType: contentType,
		Size:        req.Size,
	}

	err = s.storage.UploadFile(ctx, metadata, readSeeker)
	if err != nil {
		return "", errors.WithMessage(err, "save file")
	}
	return metadata.Filename, nil
}

func (s Files) GetFile(
	ctx context.Context,
	req domain.FileRequest,
	opt *types.RangeOption,
) (*entity.Metadata, io.Reader, error) {
	if opt != nil && (opt.End == 0 || opt.Start-opt.End > s.maxRangeRequestLength) {
		opt.End = opt.Start + s.maxRangeRequestLength
	}
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

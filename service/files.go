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
	GetFile(ctx context.Context, filename string, category string, opt *types.RangeOption) (*entity.Metadata, io.ReadSeekCloser, error)
	IsFileExist(ctx context.Context, filename string, category string) (bool, error)
	DeleteFile(ctx context.Context, filename string, category string) error
}

type Files struct {
	storage            FileStorage
	supportedFileTypes []string
}

func NewFiles(storage FileStorage, supportedFileTypes []string) Files {
	return Files{
		storage:            storage,
		supportedFileTypes: supportedFileTypes,
	}
}

func (s Files) UploadFile(ctx context.Context, req entity.UploadFileRequest) (string, error) {
	if req.ContentReader == nil {
		return "", domain.NewInvalidArgumentError("file has zero size", domain.ErrCodeFileHasZeroSize)
	}

	readSeeker, ok := req.ContentReader.(io.ReadSeeker)
	var buf []byte
	var err error

	if !ok {
		buf, err = io.ReadAll(req.ContentReader)
		if err != nil {
			return "", errors.WithMessage(err, "read content for mime detection")
		}
		readSeeker = bytes.NewReader(buf)
	}

	var contentType string
	if ok {
		contentType, err = detectMimeTypeFromSeeker(readSeeker)
	} else {
		contentType = mimetype.Detect(buf).String()
	}
	if err != nil {
		return "", errors.WithMessage(err, "detect file mime type")
	}

	if len(s.supportedFileTypes) != 0 && !slices.Contains(s.supportedFileTypes, contentType) {
		return "", domain.NewInvalidArgumentError(
			fmt.Sprintf("file type is not supported. file type: '%s'", contentType),
			domain.ErrCodeUnsupportedFileType,
		)
	}

	fileSize, err := readSeeker.Seek(0, io.SeekEnd)
	if err != nil {
		return "", err
	}
	if fileSize <= 0 {
		return "", domain.NewInvalidArgumentError("file has zero size", domain.ErrCodeFileHasZeroSize)
	}

	filename := req.Filename
	if filename == "" {
		filename = uuid.NewString()
	}

	metadata := entity.Metadata{
		Filename:    filename,
		Category:    req.Category,
		ContentType: contentType,
		Size:        fileSize,
	}

	_, err = readSeeker.Seek(0, io.SeekStart)
	if err != nil {
		return "", errors.WithMessage(err, "reset reader position")
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

func detectMimeTypeFromSeeker(rs io.ReadSeeker) (string, error) {
	peekBuf := make([]byte, 512)
	n, err := rs.Read(peekBuf)
	if err != nil && err != io.EOF {
		return "", err
	}
	_, err = rs.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	return mimetype.Detect(peekBuf[:n]).String(), nil
}

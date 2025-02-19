package controller

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Falokut/go-kit/log"

	"github.com/pkg/errors"

	"github.com/Falokut/go-kit/http/apierrors"
	"github.com/Falokut/storage_service/domain"
	"github.com/Falokut/storage_service/entity"
)

//go:generate mockgen -source=service.go -destination=mocks/service.go
type StorageService interface {
	UploadFile(ctx context.Context, req domain.UploadFileRequest) (string, error)
	GetFile(ctx context.Context, req domain.FileRequest) (*entity.File, error)
	IsFileExist(ctx context.Context, req domain.FileRequest) (bool, error)
	DeleteFile(ctx context.Context, req domain.FileRequest) error
}

type Files struct {
	service     StorageService
	maxFileSize int64
	logger      log.Logger
}

func NewFiles(
	service StorageService,
	maxFileSize int64,
) Files {
	return Files{
		service:     service,
		maxFileSize: maxFileSize,
	}
}

// UploadFile
//
//	@Tags			file
//	@Summary		Upload file
//	@Description	Загрузить файл в хранилище
//	@Accept			json
//	@Produce		json
//
//	@Param			category	path		string						true	"Категория файла"
//
//	@Param			body		body		domain.UploadFileRequest	true	"request body"
//	@Success		200			{object}	domain.UploadFileResponse
//	@Failure		400			{object}	apierrors.Error
//	@Failure		500			{object}	apierrors.Error
//	@Router			/file/{category} [POST]
func (c Files) UploadFile(ctx context.Context, req domain.UploadFileRequest) (*domain.UploadFileResponse, error) {
	filename, err := c.service.UploadFile(ctx, req)
	if err != nil {
		return nil, c.handleError(err)
	}
	return &domain.UploadFileResponse{Filename: filename}, nil
}

// GetFile
//
//	@Tags			file
//	@Summary		Get file
//	@Description	Получить файл из хранилища
//
//	@Param			category	path		string	true	"Категория файла"
//	@Param			filename	path		string	true	"Идентификатор файла"
//
//	@Success		200			{array}		byte
//	@Failure		400			{object}	apierrors.Error
//	@Failure		404			{object}	apierrors.Error
//	@Failure		500			{object}	apierrors.Error
//	@Router			/file/{category}/{filename} [GET]
func (c Files) GetFile(ctx context.Context, w http.ResponseWriter, req domain.FileRequest) error {
	file, err := c.service.GetFile(ctx, req)
	if err != nil {
		return c.handleError(err)
	}

	w.Header().Set("Content-Type", file.Metadata.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.Metadata.Size, 10))
	_, err = w.Write(file.Content)
	return errors.WithMessage(err, "write response")
}

// IsFileExist
//
//	@Tags			file
//	@Summary		Is file exist
//	@Description	Проверить наличие файла в хранилище
//	@Accept			json
//	@Produce		json
//
//	@Param			category	path		string	true	"Категория файла"
//	@Param			filename	path		string	true	"Идентификатор файла"
//
//	@Success		200			{object}	domain.FileExistResponse
//	@Failure		400			{object}	apierrors.Error
//	@Failure		404			{object}	apierrors.Error
//	@Failure		500			{object}	apierrors.Error
//	@Router			/file/{category}/{filename}/exist [GET]
func (c Files) IsFileExist(ctx context.Context, req domain.FileRequest) (*domain.FileExistResponse, error) {
	imageExist, err := c.service.IsFileExist(ctx, req)
	if err != nil {
		return nil, c.handleError(err)
	}
	return &domain.FileExistResponse{FileExist: imageExist}, nil
}

// DeleteFile
//
//	@Tags			file
//	@Summary		Delete file
//	@Description	Удалить файл из хранилища
//	@Accept			json
//	@Produce		json
//
//	@Param			category	path		string	true	"Категория файла"
//	@Param			filename	path		string	true	"Идентификатор файла"
//
//	@Success		200			{object}	any
//	@Failure		400			{object}	apierrors.Error
//	@Failure		404			{object}	apierrors.Error
//	@Failure		500			{object}	apierrors.Error
//	@Router			/file/{category}/{filename} [DELETE]
func (c Files) DeleteFile(ctx context.Context, req domain.FileRequest) error {
	return c.handleError(c.service.DeleteFile(ctx, req))
}

func (c Files) handleError(err error) error {
	if err == nil {
		return nil
	}

	invalidArgError := domain.InvalidArgumentError{}
	switch {
	case errors.Is(err, domain.ErrFileNotFound):
		return apierrors.New(
			http.StatusNotFound,
			domain.ErrCodeFileNotFound,
			domain.ErrFileNotFound.Error(),
			err,
		)
	case errors.As(err, &invalidArgError):
		return apierrors.NewBusinessError(invalidArgError.ErrCode, invalidArgError.Reason, err)
	default:
		return apierrors.NewInternalServiceError(err)
	}
}

package controller

import (
	"context"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/Falokut/go-kit/http/apierrors"
	"github.com/Falokut/go-kit/http/types"
	"github.com/Falokut/storage_service/domain"
	"github.com/Falokut/storage_service/entity"
)

//go:generate mockgen -source=service.go -destination=mocks/service.go
type StorageService interface {
	UploadFile(ctx context.Context, req entity.UploadFileRequest) (string, error)
	GetFile(ctx context.Context, req domain.FileRequest, opt *types.RangeOption) (*entity.Metadata, io.Reader, error)
	IsFileExist(ctx context.Context, req domain.FileRequest) (bool, error)
	DeleteFile(ctx context.Context, req domain.FileRequest) error
}

type Files struct {
	service StorageService
}

func NewFiles(service StorageService) Files {
	return Files{
		service: service,
	}
}

// UploadFile
//
//	@Tags			file
//	@Summary		Upload file
//	@Description	Загрузить файл в хранилище
//	@Accept			*/*
//	@Produce		*/*
//
//	@Param			category	path		string	true	"Категория файла"
//	@Param			filename	path		string	false	"имя файла"
//
//	@Param			body		body		[]byte	true	"содержимое файла"
//
//	@Success		200			{object}	domain.UploadFileResponse
//	@Failure		400			{object}	apierrors.Error
//	@Failure		500			{object}	apierrors.Error
//	@Router			/file/{category} [POST]
func (c Files) UploadFile(ctx context.Context, r *http.Request, req domain.UploadFileRequest) (*domain.UploadFileResponse, error) {
	filename, err := c.service.UploadFile(ctx,
		entity.UploadFileRequest{
			Filename:      req.Filename,
			Category:      req.Category,
			ContentReader: r.Body,
			Size:          r.ContentLength,
		})
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
func (c Files) GetFile(ctx context.Context, rangeOpt *types.RangeOption, req domain.FileRequest) (*types.FileData, error) {
	metadata, reader, err := c.service.GetFile(ctx, req, rangeOpt)
	if err != nil {
		return nil, c.handleError(err)
	}

	var partialDataInfo *types.PartialDataInfo
	if rangeOpt != nil {
		partialDataInfo = &types.PartialDataInfo{
			RangeStartByte: rangeOpt.Start,
			RangeEndByte:   rangeOpt.End,
		}
	}
	return &types.FileData{
		PartialDataInfo: partialDataInfo,
		ContentType:     metadata.ContentType,
		ContentReader:   reader,
		TotalFileSize:   metadata.Size,
	}, nil
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

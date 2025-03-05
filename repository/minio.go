package repository

import (
	"context"
	"io"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"

	"github.com/Falokut/go-kit/client/minio"
	"github.com/Falokut/go-kit/http/types"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/storage_service/domain"
	"github.com/Falokut/storage_service/entity"
)

type MinioStorage struct {
	logger log.Logger
	cli    minio_client.Client
}

func NewMinioStorage(logger log.Logger, cli minio_client.Client) MinioStorage {
	return MinioStorage{
		logger: logger,
		cli:    cli,
	}
}

func (s MinioStorage) UploadFile(ctx context.Context, metadata entity.Metadata, reader io.Reader) error {
	err := s.createBucketIfNotExist(ctx, metadata.Category)
	if err != nil {
		return errors.WithMessage(err, "create bucket if not exits")
	}

	s.logger.Info(ctx, "save file",
		log.Any("bucketName", metadata.Category),
		log.Any("filename", metadata.Filename),
	)

	putOptions := minio.PutObjectOptions{
		UserMetadata: map[string]string{
			"Name": metadata.Filename,
		},
		ContentType: metadata.ContentType,
	}
	_, err = s.cli.PutObject(ctx, metadata.Category, metadata.Filename, reader, metadata.Size, putOptions)
	if err != nil {
		return errors.WithMessage(err, "put object")
	}
	return nil
}

func (s MinioStorage) GetFile(
	ctx context.Context,
	filename string,
	category string,
	rangeOpt *types.RangeOption,
) (*entity.Metadata, io.ReadSeekCloser, error) {
	getOptions := minio.GetObjectOptions{}

	if rangeOpt != nil {
		err := getOptions.SetRange(rangeOpt.Start, rangeOpt.End)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "set range")
		}
	}

	obj, err := s.cli.GetObject(ctx, category, filename, getOptions)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "get object")
	}
	objectInfo, err := obj.Stat()
	switch {
	case minio.ToErrorResponse(err).StatusCode == http.StatusNotFound:
		return nil, nil, domain.ErrFileNotFound
	case err != nil:
		return nil, nil, errors.WithMessage(err, "get object info")
	default:
		return &entity.Metadata{
			Filename:    filename,
			Category:    category,
			ContentType: objectInfo.ContentType,
			Size:        objectInfo.Size,
		}, obj, nil
	}
}

func (s MinioStorage) IsFileExist(ctx context.Context, filename string, category string) (exist bool, err error) {
	_, err = s.cli.StatObject(ctx, category, filename, minio.StatObjectOptions{})
	switch {
	case minio.ToErrorResponse(err).StatusCode == http.StatusNotFound:
		return false, nil
	case err != nil:
		return false, errors.WithMessage(err, "stat object")
	default:
		return true, nil
	}
}

func (s MinioStorage) DeleteFile(ctx context.Context, filename string, category string) error {
	err := s.cli.RemoveObject(ctx, category, filename,
		minio.RemoveObjectOptions{ForceDelete: true})
	switch {
	case minio.ToErrorResponse(err).StatusCode == http.StatusNotFound:
		return domain.ErrFileNotFound
	case err != nil:
		return errors.WithMessage(err, "stat object")
	default:
		return nil
	}
}

func (s MinioStorage) createBucketIfNotExist(ctx context.Context, bucketName string) error {
	exists, err := s.cli.BucketExists(ctx, bucketName)
	if err != nil {
		return errors.WithMessage(err, "check is bucket exists")
	}
	if !exists {
		s.logger.Info(ctx, "creating bucket", log.Any("bucketName", bucketName))
		err = s.cli.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return errors.WithMessage(err, "make bucket")
		}
	}
	return nil
}

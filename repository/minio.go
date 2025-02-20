package repository

import (
	"bytes"
	"context"
	"fmt"
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

func (s MinioStorage) UploadFile(ctx context.Context, req entity.File) error {
	err := s.createBucketIfNotExist(ctx, req.Metadata.Category)
	if err != nil {
		return errors.WithMessage(err, "create bucket if not exits")
	}

	reader := bytes.NewReader(req.Content)
	s.logger.Info(ctx, "save file",
		log.Any("bucketName", req.Metadata.Category),
		log.Any("filename", req.Metadata.Filename),
	)

	putOptions := minio.PutObjectOptions{
		UserMetadata: map[string]string{
			"Name": req.Metadata.Filename,
		},
		ContentType: req.Metadata.ContentType,
	}
	_, err = s.cli.PutObject(ctx, req.Metadata.Category, req.Metadata.Filename, reader, req.Metadata.Size, putOptions)
	if err != nil {
		return errors.WithMessage(err, "put object")
	}
	return nil
}

func (s MinioStorage) GetFile(ctx context.Context, filename string, category string, rangeOpt *types.RangeOption) (*entity.File, error) {
	getOptions := minio.GetObjectOptions{}

	if rangeOpt != nil {
		if err := getOptions.SetRange(rangeOpt.Start, rangeOpt.End); err != nil {
			return nil, errors.WithMessage(err, "set range")
		}
		fmt.Println(rangeOpt)
	}

	obj, err := s.cli.GetObject(ctx, category, filename, getOptions)
	if err != nil {
		return nil, errors.WithMessage(err, "get object")
	}
	defer obj.Close()

	objectInfo, err := obj.Stat()
	errResp := minio.ToErrorResponse(err)
	switch {
	case errResp.StatusCode == http.StatusNotFound:
		return nil, domain.ErrFileNotFound
	case err != nil:
		return nil, errors.WithMessage(err, "get object info")
	}

	resp := &entity.File{
		Metadata: entity.Metadata{
			Filename:    filename,
			Category:    category,
			ContentType: objectInfo.ContentType,
			Size:        objectInfo.Size,
		},
	}

	if rangeOpt != nil {
		length, err := rangeOpt.Length(objectInfo.Size)
		if err != nil {
			return nil, errors.WithMessage(err, "get length")
		}
		resp.Content = make([]byte, length)
		_, err = io.ReadFull(obj, resp.Content)
		if err != nil && err != io.EOF {
			return nil, errors.WithMessage(err, "read ranged object")
		}
	} else {
		resp.Content = make([]byte, objectInfo.Size)
		_, err = io.ReadFull(obj, resp.Content)
		if err != nil && err != io.EOF {
			return nil, errors.WithMessage(err, "read full object")
		}
	}

	return resp, nil
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

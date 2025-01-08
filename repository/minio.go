package repository

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"io"
	"net/http"

	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/storage_service/domain"
	"github.com/Falokut/storage_service/entity"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	logger            log.Logger
	storage           *minio.Client
	uploadFileThreads uint
}

type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Secure          bool
	Token           string
}

func NewMinio(cfg MinioConfig) (*minio.Client, error) {
	return minio.New(cfg.Endpoint,
		&minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.Token),
			Secure: cfg.Secure,
		},
	)
}

func NewMinioStorage(logger log.Logger, storage *minio.Client, uploadFileThreads uint) MinioStorage {
	return MinioStorage{
		logger:            logger,
		storage:           storage,
		uploadFileThreads: uploadFileThreads,
	}
}

func (s MinioStorage) UploadFile(ctx context.Context, req entity.File) error {
	err := s.createBucketIfNotExist(ctx, req.Category)
	if err != nil {
		return errors.WithMessage(err, "create bucket if not exits")
	}

	reader := bytes.NewReader(req.Content)
	s.logger.Info(ctx, "save file",
		log.Any("bucketName", req.Category),
		log.Any("filename", req.Filename),
	)

	putOptions := minio.PutObjectOptions{
		UserMetadata: map[string]string{
			"Name": req.Filename,
		},
		ContentType:           req.ContentType,
		ConcurrentStreamParts: s.uploadFileThreads > 1,
		NumThreads:            s.uploadFileThreads,
	}
	_, err = s.storage.PutObject(ctx, req.Category, req.Filename, reader, req.Size, putOptions)
	if err != nil {
		return errors.WithMessage(err, "put object")
	}
	return nil
}

func (s MinioStorage) GetFile(ctx context.Context, filename string, category string) (*entity.File, error) {
	obj, err := s.storage.GetObject(ctx,
		category, filename, minio.GetObjectOptions{})
	errResp := minio.ToErrorResponse(err)
	switch {
	case errResp.StatusCode == http.StatusNotFound:
		return nil, domain.ErrFileNotFound
	case err != nil:
		return nil, errors.WithMessage(err, "get object")
	}
	defer obj.Close()

	objectInfo, err := obj.Stat()
	if err != nil {
		return nil, errors.WithMessage(err, "object stat")
	}

	content := make([]byte, objectInfo.Size)
	_, err = obj.Read(content)
	if err != nil && err != io.EOF {
		return nil, errors.WithMessage(err, "read object")
	}

	return &entity.File{
		Filename:    filename,
		Category:    category,
		ContentType: objectInfo.ContentType,
		Content:     content,
		Size:        objectInfo.Size,
	}, nil
}

func (s MinioStorage) IsFileExist(ctx context.Context, filename string, category string) (exist bool, err error) {
	_, err = s.storage.StatObject(ctx, category, filename, minio.StatObjectOptions{})
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
	err := s.storage.RemoveObject(ctx, category, filename,
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
	exists, err := s.storage.BucketExists(ctx, bucketName)
	if err != nil {
		return errors.WithMessage(err, "check is bucket exists")
	}
	if !exists {
		s.logger.Info(ctx, "creating bucket", log.Any("bucketName", bucketName))
		err = s.storage.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return errors.WithMessage(err, "make bucket")
		}
	}
	return nil
}

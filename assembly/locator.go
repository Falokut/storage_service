package assembly

import (
	"context"
	"strings"
	"time"

	"github.com/Falokut/go-kit/healthcheck"
	"github.com/Falokut/go-kit/http/endpoint"
	"github.com/Falokut/go-kit/http/router"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/storage_service/conf"
	"github.com/Falokut/storage_service/controller"
	"github.com/Falokut/storage_service/repository"
	"github.com/Falokut/storage_service/routes"
	"github.com/Falokut/storage_service/service"
	"github.com/pkg/errors"
)

const (
	mb = 8 << 20
)

type Config struct {
	Mux *router.Router
}

func Locator(_ context.Context,
	logger log.Logger,
	cfg conf.LocalConfig,
	healthcheckManager healthcheck.Manager,
) (Config, error) {
	storageMode := strings.ToUpper(cfg.StorageMode)
	filesStorage, err := getFilesStorage(storageMode, cfg, logger, healthcheckManager)
	if err != nil {
		return Config{}, errors.WithMessage(err, "get files storage")
	}

	filesService := service.NewFiles(filesStorage,
		cfg.MaxImageSizeMb*mb, cfg.SupportedFileTypes)
	filesController := controller.NewFiles(filesService, cfg.MaxImageSizeMb*mb)
	c := routes.Router{
		Files: filesController,
	}
	defaultWrapper := endpoint.DefaultWrapper(logger)
	mux := c.InitRoutes(defaultWrapper)
	return Config{
		Mux: mux,
	}, nil
}

func getFilesStorage(
	storageMode string,
	cfg conf.LocalConfig,
	logger log.Logger,
	healthcheckManager healthcheck.Manager,
) (service.FileStorage, error) {
	switch storageMode {
	case "MINIO":
		minioStorage, err := repository.NewMinio(repository.MinioConfig{
			Endpoint:        cfg.MinioConfig.Endpoint,
			AccessKeyID:     cfg.MinioConfig.AccessKeyID,
			SecretAccessKey: cfg.MinioConfig.SecretAccessKey,
			Secure:          cfg.MinioConfig.Secure,
			Token:           cfg.MinioConfig.Token,
		})
		if err != nil {
			return nil, errors.WithMessage(err, "new minio")
		}
		_, err = minioStorage.HealthCheck(time.Second * 5)
		if err != nil {
			return nil, errors.WithMessage(err, "init healthcheck minio")
		}
		healthcheckManager.Register("minio-storage", func(ctx context.Context) error {
			if minioStorage.IsOnline() {
				return nil
			}
			return errors.New("minio offline")
		})
		return repository.NewMinioStorage(logger, minioStorage, cfg.MinioConfig.UploadFileThreads), nil
	case "LOCAL":
		return repository.NewLocalStorage(cfg.BaseLocalStoragePath), nil
	}
	return nil, errors.New("unknown storage type")
}

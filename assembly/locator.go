package assembly

import (
	"context"

	"github.com/Falokut/go-kit/client/minio"
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
	kb = 8 << 10
	mb = kb << 10
)

type Config struct {
	Mux *router.Router
}

func Locator(_ context.Context,
	logger log.Logger,
	cfg conf.LocalConfig,
	healthcheckManager healthcheck.Manager,
) (Config, error) {
	minioCli, err := minio_client.NewMinio(cfg.MinioConfig)
	if err != nil {
		return Config{}, errors.WithMessage(err, "new minio")
	}
	healthcheckManager.Register("minio-storage", minioCli.HealthCheck)

	filesStorage := repository.NewMinioStorage(logger, minioCli)

	filesService := service.NewFiles(filesStorage,
		cfg.MaxImageSizeMb*mb, cfg.MaxRangeRequestLength*mb, cfg.SupportedFileTypes)
	filesController := controller.NewFiles(filesService,
		cfg.MaxImageSizeMb*mb)
	c := routes.Router{
		Files: filesController,
	}
	defaultWrapper := endpoint.DefaultWrapper(logger)
	mux := c.InitRoutes(defaultWrapper)
	return Config{
		Mux: mux,
	}, nil
}

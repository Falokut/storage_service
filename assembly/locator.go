package assembly

import (
	"context"

	minio_client "github.com/Falokut/go-kit/client/minio"
	"github.com/Falokut/go-kit/healthcheck"
	http2 "github.com/Falokut/go-kit/http"
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
	kb = 1 << 10
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
		cfg.SupportedFileTypes,
	)
	files := controller.NewFiles(filesService)
	c := routes.Router{Files: files}

	defaultWrapper := newWrapper(logger, cfg.MaxFileSizeMb*mb)
	mux := c.InitRoutes(defaultWrapper)
	return Config{
		Mux: mux,
	}, nil
}

func newWrapper(logger log.Logger, maxRequestBody int64) endpoint.Wrapper {
	wrapper := endpoint.DefaultWrapper(logger, nil)
	wrapper.Middlewares = []http2.Middleware{
		endpoint.MaxRequestBodySize(maxRequestBody),
		endpoint.RequestId(),
		http2.Middleware(endpoint.Log(logger, false, false)),
		endpoint.ErrorHandler(logger),
		endpoint.Recovery(),
	}
	return wrapper
}

package assembly

import (
	"context"
	"fmt"

	"github.com/Falokut/go-kit/app"
	"github.com/Falokut/go-kit/config"
	"github.com/Falokut/go-kit/healthcheck"
	"github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/storage_service/conf"
	"github.com/pkg/errors"
)

type Assembly struct {
	logger             log.Logger
	server             *http.Server
	healthcheckManager healthcheck.Manager
	cfg                conf.LocalConfig
}

func New(ctx context.Context, logger log.Logger) (*Assembly, error) {
	var cfg conf.LocalConfig
	err := config.Read(&cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "read local config")
	}
	server := http.NewServer(logger)
	listenHealthcheckPort := cfg.HealthcheckPort
	if listenHealthcheckPort == 0 {
		listenHealthcheckPort = cfg.Listen.Port + 1
	}
	healthcheckManager := healthcheck.NewHealthManager(logger, fmt.Sprint(listenHealthcheckPort))

	locatorCfg, err := Locator(ctx, logger, cfg, healthcheckManager)
	if err != nil {
		return nil, errors.WithMessage(err, "init locator")
	}

	server.Upgrade(locatorCfg.Mux)
	return &Assembly{
		logger:             logger,
		server:             server,
		healthcheckManager: healthcheckManager,
		cfg:                cfg,
	}, nil
}

func (a *Assembly) Runners() []app.RunnerFunc {
	return []app.RunnerFunc{
		func(ctx context.Context) error {
			a.logger.Debug(ctx, "run http server", log.Any("host", a.cfg.Listen.Host), log.Any("port", a.cfg.Listen.Port))
			err := a.server.ListenAndServe(a.cfg.Listen.GetAddress())
			if err != nil {
				a.logger.Debug(ctx,
					"error while listen and serve http server",
					log.Any("host", a.cfg.Listen.Host),
					log.Any("port", a.cfg.Listen.Port),
					log.Any("error", err.Error()),
				)
				return err
			}
			return nil
		},
		func(context.Context) error {
			return a.healthcheckManager.RunHealthcheckEndpoint()
		},
	}
}

func (a *Assembly) Closers() []app.CloserFunc {
	return []app.CloserFunc{
		func(ctx context.Context) error {
			return a.server.Shutdown(ctx)
		},
	}
}

package assembly

import (
	"context"

	"github.com/Falokut/go-kit/app"
	"github.com/Falokut/go-kit/config"
	"github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/storage_service/conf"
	"github.com/pkg/errors"
)

type Assembly struct {
	logger log.Logger
	server *http.Server
	cfg    conf.LocalConfig
}

func New(ctx context.Context, logger log.Logger) (*Assembly, error) {
	var cfg conf.LocalConfig
	err := config.Read(&cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "read local config")
	}
	server := http.NewServer(logger)
	locatorCfg, err := Locator(ctx, logger, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "init locator")
	}
	server.Upgrade(locatorCfg.Mux)
	return &Assembly{
		logger: logger,
		server: server,
		cfg:    cfg,
	}, nil
}

func (a *Assembly) Runners() []app.RunnerFunc {
	return []app.RunnerFunc{
		func(context.Context) error {
			return a.server.ListenAndServe(a.cfg.Listen.GetAddress())
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

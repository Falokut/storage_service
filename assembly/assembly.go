package assembly

import (
	"context"
	"storage-service/conf"
	"storage-service/service/pending"

	"github.com/Falokut/go-kit/cluster"
	"github.com/Falokut/go-kit/dbx"
	"github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/miniox"
	"github.com/Falokut/go-kit/remote"

	"github.com/Falokut/go-kit/app"
	"github.com/Falokut/go-kit/bootstrap"
	"github.com/Falokut/go-kit/db"
	"github.com/Falokut/go-kit/log"
	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
)

type Assembly struct {
	boot     *bootstrap.Bootstrap
	db       *dbx.Client
	minioCli *miniox.Client
	server   *http.Server
	logger   *log.Adapter
}

func New(boot *bootstrap.Bootstrap) (*Assembly, error) {
	logger := boot.App.Logger()

	db := dbx.New(logger, db.WithMigrationRunner(boot.MigrationsDir, logger))
	boot.HealthcheckRegistry.Register("db", db)

	minioCli := miniox.New(logger)
	boot.HealthcheckRegistry.Register("minio", minioCli)

	server := http.NewServer(logger)
	return &Assembly{
		boot:     boot,
		db:       db,
		minioCli: minioCli,
		server:   server,
		logger:   logger,
	}, nil
}

func (a *Assembly) ReceiveConfig(shortCtx context.Context, remoteConfig []byte) error {
	newCfg, _, err := remote.Upgrade[conf.Remote](a.boot.RemoteConfig, remoteConfig)
	if err != nil {
		a.boot.Fatal(errors.WithMessage(err, "upgrade remote config"))
	}
	a.logger.SetLogLevel(newCfg.LogLevel)

	err = a.db.Upgrade(shortCtx, newCfg.DB)
	if err != nil {
		a.boot.Fatal(errors.WithMessage(err, "upgrade db"))
	}

	err = a.minioCli.Upgrade(a.boot.App.Context(), newCfg.Minio, miniox.OptionsFromConfig(newCfg.Minio)...)
	if err != nil {
		a.boot.Fatal(errors.WithMessage(err, "upgrade minio"))
	}

	pgDb, _ := a.db.DB()
	bgjobCli := bgjob.NewClient(bgjob.NewPgStore(pgDb.DB.DB))
	err = pending.EnqueueSeedJob(shortCtx, bgjobCli)
	if err != nil {
		a.boot.Fatal(errors.WithMessage(err, "enqueu pending job"))
	}

	minioCli, err := a.minioCli.Client()
	if err != nil {
		a.boot.Fatal(errors.WithMessage(err, "get minio client"))
	}

	locator := NewLocator(a.db, bgjobCli, minioCli, a.logger)
	cfg, err := locator.LocatorConfig(shortCtx, newCfg)
	if err != nil {
		a.boot.Fatal(errors.WithMessage(err, "locator config"))
	}

	for _, worker := range cfg.Workers {
		worker.Run(a.boot.App.Context())
	}

	a.server.Upgrade(cfg.HttpRouter)
	return nil
}

func (a *Assembly) Runners() []app.Runner {
	eventHandler := cluster.NewEventHandler().
		RemoteConfigReceiver(a)
	return []app.Runner{
		app.RunnerFunc(func(_ context.Context) error {
			err := a.server.ListenAndServe(a.boot.BindingAddress)
			if err != nil {
				return errors.WithMessage(err, "listen and serve http server")
			}
			return nil
		}),
		app.RunnerFunc(func(ctx context.Context) error {
			err := a.boot.ClusterCli.Run(ctx, eventHandler)
			if err != nil {
				return errors.WithMessage(err, "run cluster client")
			}
			return nil
		}),
	}
}

func (a *Assembly) Closers() []app.Closer {
	return []app.Closer{
		app.CloserFunc(func(_ context.Context) error {
			return a.db.Close()
		}),
		app.CloserFunc(func(_ context.Context) error {
			return nil
		}),
		app.CloserFunc(func(ctx context.Context) error {
			return a.server.Shutdown(ctx)
		}),
	}
}

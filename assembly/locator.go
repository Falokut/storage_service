package assembly

import (
	"context"
	"time"

	"storage-service/conf"
	"storage-service/controller"
	"storage-service/repository"
	"storage-service/routes"
	"storage-service/service"
	"storage-service/service/pending"
	"storage-service/transaction"

	"github.com/Falokut/go-kit/db"
	http2 "github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/http/endpoint"
	"github.com/Falokut/go-kit/http/endpoint/hlog"
	"github.com/Falokut/go-kit/http/router"
	"github.com/Falokut/go-kit/log"
	"github.com/minio/minio-go/v7"
	"github.com/txix-open/bgjob"
)

const (
	kb = 1 << 10
	mb = kb << 10
)

type DB interface {
	db.DB
	db.Transactional
}

type Locator struct {
	logger   *log.Adapter
	db       DB
	bgJobCli *bgjob.Client
	minioCli *minio.Client
}

func NewLocator(
	db DB,
	bgJobCli *bgjob.Client,
	minioCli *minio.Client,
	logger *log.Adapter,
) Locator {
	return Locator{
		db:       db,
		bgJobCli: bgJobCli,
		minioCli: minioCli,
		logger:   logger,
	}
}

type Config struct {
	HttpRouter *router.Router
	Workers    []*bgjob.Worker
}

func (l Locator) LocatorConfig(ctx context.Context, cfg conf.Remote) (*Config, error) {
	txRunner := transaction.NewManager(l.db)
	filesStorage := repository.NewMinioStorage(l.logger, l.minioCli)

	pendingRepo := repository.NewPending(l.db)
	pendingFileLifetime := time.Duration(cfg.Pending.FileLifetimeInMin) * time.Minute
	pendingService := pending.NewPending(
		txRunner,
		filesStorage,
		pendingRepo,
		pendingFileLifetime,
		cfg.Pending.MaxFilesToDelete,
	)
	filesService := service.NewFiles(
		filesStorage,
		cfg.SupportedFileTypes,
		pendingService,
	)
	files := controller.NewFiles(filesService)
	c := routes.Router{
		Files: files,
	}

	defaultWrapper := newWrapper(l.logger, cfg.MaxFileSizeMb*mb)
	mux := c.Handler(defaultWrapper)
	observer := service.NewObserver(l.logger)

	pendingFileController := controller.NewPendingWorker(pendingService)

	return &Config{
		HttpRouter: mux,
		Workers: []*bgjob.Worker{
			bgjob.NewWorker(
				l.bgJobCli,
				pending.WorkerQueueName,
				pendingFileController,
				bgjob.WithPollInterval(5*time.Second), // nolint:mnd
				bgjob.WithObserver(observer),
			),
		},
	}, nil
}

func newWrapper(logger log.Logger, maxRequestBody int64) endpoint.Wrapper {
	wrapper := endpoint.DefaultWrapper(logger, nil)
	wrapper.Middlewares = []http2.Middleware{
		endpoint.MaxRequestBodySize(maxRequestBody),
		endpoint.RequestId(),
		http2.Middleware(hlog.Log(logger, false)),
		endpoint.ErrorHandler(logger),
		endpoint.Recovery(),
	}
	return wrapper
}

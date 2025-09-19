package controller

import (
	"context"
	"time"

	"github.com/txix-open/bgjob"
)

const (
	defaultRetryTime = time.Minute * 5
)

type PendingService interface {
	ProcessPendingFiles(ctx context.Context) error
}

type Pending struct {
	service PendingService
}

func NewPendingWorker(service PendingService) Pending {
	return Pending{
		service: service,
	}
}

func (c Pending) Handle(ctx context.Context, job bgjob.Job) bgjob.Result {
	err := c.service.ProcessPendingFiles(ctx)
	if err != nil {
		return bgjob.Retry(defaultRetryTime, err)
	}
	return bgjob.Reschedule(defaultRetryTime)
}

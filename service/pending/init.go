package pending

import (
	"context"

	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
)

const WorkerQueueName = "pending_files"

func EnqueueSeedJob(ctx context.Context, client *bgjob.Client) error {
	err := client.Enqueue(ctx, bgjob.EnqueueRequest{
		Id:    "pending",
		Queue: WorkerQueueName,
		Type:  "pending",
	})
	if err != nil && !errors.Is(err, bgjob.ErrJobAlreadyExist) {
		return errors.WithMessage(err, "enqueue job")
	}

	return nil
}

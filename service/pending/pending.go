package pending

import (
	"context"
	"storage-service/domain"
	"storage-service/entity"
	"time"

	"github.com/pkg/errors"
)

type PendingTxRunner interface {
	DeletePendingFilesTx(ctx context.Context, tx func(ctx context.Context, tx PendingFilesTx) error) error
}

type PendingFilesTx interface {
	DeletePendingFiles(ctx context.Context, timeAgo time.Time, maxFiles int) ([]entity.FileToDelete, error)
	DeletePendingFile(ctx context.Context, filename string, category string) error
}

type PendingFileRepo interface {
	DeleteFile(ctx context.Context, filename string, category string) error
}

type PendingRepo interface {
	InsertPendingFile(ctx context.Context, filename string, category string, createdAt time.Time) error
	DeletePendingFile(ctx context.Context, filename string, category string) error
}

type Pending struct {
	txRunner            PendingTxRunner
	repo                PendingFileRepo
	pendingRepo         PendingRepo
	pendingFileLifetime time.Duration
	maxDeletedFiles     int
}

func NewPending(
	txRunner PendingTxRunner,
	repo PendingFileRepo,
	pendingRepo PendingRepo,
	pendingFileLifetime time.Duration,
	maxDeleteFiles int,
) Pending {
	return Pending{
		txRunner: txRunner,
		repo:     repo,
	}
}

func (s Pending) Enqueue(ctx context.Context, fileName string, category string) error {
	err := s.pendingRepo.InsertPendingFile(ctx, fileName, category, time.Now().UTC())
	if err != nil {
		return errors.WithMessage(err, "insert pending file")
	}
	return nil
}

func (s Pending) Commit(ctx context.Context, fileName string, category string) error {
	err := s.pendingRepo.DeletePendingFile(ctx, fileName, category)
	if err != nil {
		return errors.WithMessage(err, "delete pending file")
	}
	return nil
}

func (s Pending) Rollback(ctx context.Context, fileName string, category string) error {
	err := s.txRunner.DeletePendingFilesTx(ctx, func(ctx context.Context, tx PendingFilesTx) error {
		err := tx.DeletePendingFile(ctx, fileName, category)
		if err != nil {
			return errors.WithMessage(err, "delete pending files")
		}
		err = s.processPendingFile(ctx, entity.FileToDelete{Filename: fileName, Category: category})
		if err != nil {
			return errors.WithMessage(err, "process pendng file")
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "delete pending files tx")
	}
	return nil
}

func (s Pending) ProcessPendingFiles(ctx context.Context) error {
	timeAgo := time.Now().UTC().Add(-s.pendingFileLifetime)
	err := s.txRunner.DeletePendingFilesTx(ctx, func(ctx context.Context, tx PendingFilesTx) error {
		files, err := tx.DeletePendingFiles(ctx, timeAgo, s.maxDeletedFiles)
		if err != nil {
			return errors.WithMessage(err, "delete pending files")
		}
		for _, file := range files {
			err = s.processPendingFile(ctx, file)
			if err != nil {
				return errors.WithMessage(err, "process pendng file")
			}
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "delete pending files tx")
	}
	return nil
}

func (s Pending) processPendingFile(ctx context.Context, file entity.FileToDelete) error {
	err := s.repo.DeleteFile(ctx, file.Filename, file.Category)
	switch {
	case errors.Is(err, domain.ErrFileNotFound):
		return nil
	case err != nil:
		return errors.WithMessagef(err, "delete file with name '%s' and category '%s'", file.Filename, file.Category)
	default:
		return nil
	}
}

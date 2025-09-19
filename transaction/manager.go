package transaction

import (
	"context"
	"storage-service/repository"

	"storage-service/service/pending"

	"github.com/Falokut/go-kit/db"
)

type Manager struct {
	db db.Transactional
}

func NewManager(db db.Transactional) *Manager {
	return &Manager{db: db}
}

type pendingTransaction struct {
	repository.Pending
}

func (m *Manager) DeletePendingFilesTx(ctx context.Context, txRequest func(ctx context.Context, tx pending.PendingFilesTx) error) error {
	return m.db.RunInTransaction(
		ctx,
		func(ctx context.Context, tx *db.Tx) error {
			pending := repository.NewPending(tx)
			return txRequest(ctx, pendingTransaction{pending})
		},
	)
}

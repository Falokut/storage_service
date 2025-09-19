package repository

import (
	"context"
	"storage-service/entity"
	"time"

	"github.com/Falokut/go-kit/db"
	"github.com/pkg/errors"
)

type Pending struct {
	db db.DB
}

func NewPending(db db.DB) Pending {
	return Pending{
		db: db,
	}
}

func (r Pending) DeletePendingFiles(ctx context.Context, timeAgo time.Time, maxFiles int) ([]entity.FileToDelete, error) {
	files := []entity.FileToDelete{}

	query := `
		WITH deleted AS (
			SELECT filename, category
			FROM pending_files
			WHERE created_at <= $1
			ORDER BY created_at
			LIMIT $2
		)
		DELETE FROM pending_files p
		USING deleted d
		WHERE p.filename = d.filename AND p.category = d.category
		RETURNING p.filename, p.category
	`

	err := r.db.Select(ctx, &files, query, timeAgo, maxFiles)
	if err != nil {
		return nil, errors.WithMessagef(err, "exec query: %s", query)
	}

	return files, nil
}

func (r Pending) DeletePendingFile(ctx context.Context, filename string, category string) error {
	query := `
		DELETE FROM pending_files
		WHERE filename = $1 AND category = $2
	`
	_, err := r.db.Exec(ctx, query, filename, category)
	if err != nil {
		return errors.WithMessagef(err, "exec query: %s", query)
	}
	return nil
}

func (r Pending) InsertPendingFile(ctx context.Context, filename string, category string, createdAt time.Time) error {
	query := `
		INSERT INTO pending_files (filename, category, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (filename, category) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, filename, category, createdAt)
	if err != nil {
		return errors.WithMessagef(err, "exec query: %s", query)
	}
	return nil
}

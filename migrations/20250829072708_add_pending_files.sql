-- +goose Up
CREATE TABLE pending_files (
    filename TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY(filename, category)
);

-- +goose Down
DROP TABLE pending_files;
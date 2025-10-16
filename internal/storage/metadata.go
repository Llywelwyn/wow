package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sqlite3 "github.com/mattn/go-sqlite3"

	"github.com/llywelwyn/wow/internal/model"
)

// ErrMetadataNotFound indicates there is no row for the requested key.
var ErrMetadataNotFound = errors.New("metadata not found")

// ErrMetadataDuplicate indicates metadata already exists for the key.
var ErrMetadataDuplicate = errors.New("metadata already exists")

// InsertMetadata inserts a new metadata row for the provided snippet key.
func InsertMetadata(ctx context.Context, db *sql.DB, meta model.Metadata) error {
	const query = `
INSERT INTO snippets (key, type, created, modified, description, tags)
VALUES (?, ?, ?, ?, ?, ?)
`
	_, err := db.ExecContext(ctx, query, meta.Key, meta.Type, meta.Created.UTC(), meta.Modified.UTC(), meta.Description, meta.Tags)
	if err != nil {
		if sqliteIsUniqueError(err) {
			return ErrMetadataDuplicate
		}
		return fmt.Errorf("insert metadata: %w", err)
	}
	return nil
}

// GetMetadata retrieves metadata for the provided snippet key.
func GetMetadata(ctx context.Context, db *sql.DB, key string) (model.Metadata, error) {
	const query = `
SELECT key, type, created, modified, description, tags
FROM snippets
WHERE key = ?
`
	var meta model.Metadata
	err := db.QueryRowContext(ctx, query, key).Scan(
		&meta.Key,
		&meta.Type,
		&meta.Created,
		&meta.Modified,
		&meta.Description,
		&meta.Tags,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Metadata{}, ErrMetadataNotFound
	}
	if err != nil {
		return model.Metadata{}, fmt.Errorf("get metadata: %w", err)
	}
	return meta, nil
}

func sqliteIsUniqueError(err error) bool {
	var se sqlite3.Error
	if !errors.As(err, &se) {
		return false
	}
	return se.Code == sqlite3.ErrConstraint && se.ExtendedCode == sqlite3.ErrConstraintPrimaryKey
}

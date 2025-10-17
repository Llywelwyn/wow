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

// ListMetadata retrieves all metadata rows ordered from newest to oldest.
func ListMetadata(ctx context.Context, db *sql.DB) ([]model.Metadata, error) {
	const query = `
SELECT key, type, created, modified, description, tags
FROM snippets
ORDER BY created DESC
`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list metadata: %w", err)
	}
	defer rows.Close()

	var result []model.Metadata
	for rows.Next() {
		var meta model.Metadata
		if err := rows.Scan(
			&meta.Key,
			&meta.Type,
			&meta.Created,
			&meta.Modified,
			&meta.Description,
			&meta.Tags,
		); err != nil {
			return nil, fmt.Errorf("scan metadata row: %w", err)
		}
		result = append(result, meta)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate metadata: %w", err)
	}

	return result, nil
}

// DeleteMetadata removes the metadata row for the provided key.
func DeleteMetadata(ctx context.Context, db *sql.DB, key string) error {
	const query = `
DELETE FROM snippets
WHERE key = ?
`
	res, err := db.ExecContext(ctx, query, key)
	if err != nil {
		return fmt.Errorf("delete metadata: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete metadata: rows affected: %w", err)
	}
	if count == 0 {
		return ErrMetadataNotFound
	}
	return nil
}

func sqliteIsUniqueError(err error) bool {
	var se sqlite3.Error
	if !errors.As(err, &se) {
		return false
	}
	return se.Code == sqlite3.ErrConstraint && se.ExtendedCode == sqlite3.ErrConstraintPrimaryKey
}

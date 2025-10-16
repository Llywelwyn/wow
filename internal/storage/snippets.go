package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sqlite3 "github.com/mattn/go-sqlite3"

	"github.com/llywelwyn/wow/internal/model"
)

var (
	ErrSnippetNotFound  = errors.New("snippet not found")
	ErrSnippetDuplicate = errors.New("snippet already exists")
)

// InsertSnippet saves metadata for a new snippet.
func InsertSnippet(ctx context.Context, db *sql.DB, sn model.Snippet) error {
	const query = `
INSERT INTO snippets (key, type, created, modified, description, tags)
VALUES (?, ?, ?, ?, ?, ?)
`
	_, err := db.ExecContext(ctx, query, sn.Key, sn.Type, sn.Created.UTC(), sn.Modified.UTC(), sn.Description, sn.Tags)
	if err != nil {
		if sqliteIsUniqueError(err) {
			return ErrSnippetDuplicate
		}
		return fmt.Errorf("insert snippet metadata: %w", err)
	}
	return nil
}

// GetSnippet fetches metadata for the given key.
func GetSnippet(ctx context.Context, db *sql.DB, key string) (model.Snippet, error) {
	const query = `
SELECT key, type, created, modified, description, tags
FROM snippets
WHERE key = ?
`
	var sn model.Snippet
	err := db.QueryRowContext(ctx, query, key).Scan(
		&sn.Key,
		&sn.Type,
		&sn.Created,
		&sn.Modified,
		&sn.Description,
		&sn.Tags,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Snippet{}, ErrSnippetNotFound
	}
	if err != nil {
		return model.Snippet{}, fmt.Errorf("get snippet metadata: %w", err)
	}
	return sn, nil
}

func sqliteIsUniqueError(err error) bool {
	var se sqlite3.Error
	if !errors.As(err, &se) {
		return false
	}
	return se.Code == sqlite3.ErrConstraint && se.ExtendedCode == sqlite3.ErrConstraintPrimaryKey
}

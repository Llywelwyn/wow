package core

import (
	"context"
	"database/sql"
	"errors"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/storage"
)

// Remover deletes snippet content and metadata.
type Remover struct {
	BaseDir string
	DB      *sql.DB
}

// Remove deletes the snippet identified by key, returning ErrMetadataNotFound when absent.
func (r *Remover) Remove(ctx context.Context, rawKey string) error {
	if r.DB == nil {
		return errors.New("remover misconfigured")
	}

	normalized, err := key.Normalize(rawKey)
	if err != nil {
		return err
	}

	if err := storage.DeleteMetadata(ctx, r.DB, normalized); err != nil {
		return err
	}

	path, err := key.ResolvePath(r.BaseDir, normalized)
	if err != nil {
		return err
	}

	if err := storage.Delete(path); err != nil && !errors.Is(err, storage.ErrNotFound) {
		return err
	}

	return nil
}

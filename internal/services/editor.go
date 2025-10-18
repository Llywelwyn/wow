package services

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/model"
	"github.com/llywelwyn/wow/internal/storage"
)

// Editor orchestrates edit operations for existing snippets.
type Editor struct {
	BaseDir string
	DB      *sql.DB
	Now     func() time.Time
	Open    func(context.Context, string) error
}

// Edit opens the snippet for modification and refreshes metadata when changed.
func (e *Editor) Edit(ctx context.Context, rawKey string) (model.Metadata, error) {
	if e.DB == nil || e.Now == nil || e.Open == nil {
		return model.Metadata{}, errors.New("editor misconfigured")
	}

	normalized, err := key.Normalize(rawKey)
	if err != nil {
		return model.Metadata{}, err
	}

	meta, err := storage.GetMetadata(ctx, e.DB, normalized)
	if err != nil {
		return model.Metadata{}, err
	}

	path, err := key.ResolvePath(e.BaseDir, normalized)
	if err != nil {
		return model.Metadata{}, err
	}

	before, err := os.Stat(path)
	if err != nil {
		return model.Metadata{}, err
	}

	if err := e.Open(ctx, path); err != nil {
		return model.Metadata{}, err
	}

	after, err := os.Stat(path)
	if err != nil {
		return model.Metadata{}, err
	}

	changed := after.ModTime() != before.ModTime() || after.Size() != before.Size()
	if !changed {
		return meta, nil
	}

	data, err := storage.Read(path)
	if err != nil {
		return model.Metadata{}, err
	}

	meta.Type = detectType(data)
	meta.Modified = e.Now().UTC()

	if err := storage.UpdateMetadata(ctx, e.DB, meta); err != nil {
		return model.Metadata{}, err
	}

	return meta, nil
}

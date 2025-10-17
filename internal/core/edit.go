package core

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

// Edit opens the snippet for modification
// and refreshes metadata when changed.
func (e *Editor) Edit(ctx context.Context, rawKey string) (meta model.Metadata, err error) {
	// Validate dependencies
	if e.DB == nil || e.Now == nil || e.Open == nil || strings.TrimSpace(e.BaseDir) == "" {
		return model.Metadata{}, errors.New("editor misconfigured")
	}

	// Get current metadata
	meta, err = e.getMetadata(ctx, rawKey)
	if err != nil {
		return model.Metadata{}, err
	}

	// Resolve file path
	path, err := key.ResolvePath(e.BaseDir, rawKey)
	if err != nil {
		return model.Metadata{}, err
	}

	// Stat before editing
	before, err := os.Stat(path)
	if err != nil {
		return model.Metadata{}, err
	}

	// Open file for editing
	if err = e.Open(ctx, path); err != nil {
		return model.Metadata{}, err
	}

	// Stat after editing
	after, err := os.Stat(path)
	if err != nil {
		return model.Metadata{}, err
	}

	// If file unchanged, return original metadata
	if !e.fileChanged(before, after) {
		return meta, nil
	}

	// Read updated file data
	data, err := storage.Read(path)
	if err != nil {
		return model.Metadata{}, err
	}

	// Update and return new metadata
	return e.updateMetadata(data, ctx, meta)
}

// getMetadata normalizes the raw key given
// and fetches associated metadata from storage.
func (e *Editor) getMetadata(ctx context.Context, rawKey string) (model.Metadata, error) {
	normalizedKey, err := key.Normalize(rawKey)
	if err != nil {
		return model.Metadata{}, err
	}
	metadata, err := storage.GetMetadata(ctx, e.DB, normalizedKey)
	if err != nil {
		return model.Metadata{}, err
	}
	return metadata, nil
}

// fileChanged returns true if the file's mod time or size has changed.
func (e *Editor) fileChanged(before, after os.FileInfo) bool {
	return after.ModTime() != before.ModTime() || after.Size() != before.Size()
}

// updateMetadata updates Type and Modified fields.
// Type is detected from data content.
// Modified is set to the current time.
func (e *Editor) updateMetadata(data []byte, ctx context.Context, metadata model.Metadata) (model.Metadata, error) {
	metadata.Type = detectType(data)
	metadata.Modified = e.Now().UTC()
	if err := storage.UpdateMetadata(ctx, e.DB, metadata); err != nil {
		return model.Metadata{}, err
	}

	return metadata, nil
}

package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/model"
	"github.com/llywelwyn/wow/internal/storage"
)

// Metadata manages snippet metadata updates.
type Metadata struct {
	DB  *sql.DB
	Now func() time.Time
}

// UpdateTags applies additions/removals to the existing tag set.
func (m *Metadata) UpdateTags(ctx context.Context, rawKey string, add, remove []string) (model.Metadata, error) {
	if m.DB == nil || m.Now == nil {
		return model.Metadata{}, errors.New("metadata service misconfigured")
	}

	normalized, err := key.Normalize(rawKey)
	if err != nil {
		return model.Metadata{}, err
	}

	meta, err := storage.GetMetadata(ctx, m.DB, normalized)
	if err != nil {
		return model.Metadata{}, err
	}

	updated := MergeTags(meta.Tags, add, remove)
	if updated == meta.Tags {
		return meta, nil
	}

	meta.Tags = updated
	meta.Modified = m.Now().UTC()

	if err := storage.UpdateMetadata(ctx, m.DB, meta); err != nil {
		return model.Metadata{}, err
	}

	return meta, nil
}

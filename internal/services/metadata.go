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

// TagUpdateResult reports the outcome of updating snippet tags.
type TagUpdateResult struct {
	Metadata model.Metadata
	Added    []string
	Removed  []string
}

// UpdateTags applies additions/removals to the existing tag set.
func (m *Metadata) UpdateTags(ctx context.Context, rawKey string, add, remove []string) (TagUpdateResult, error) {
	if m.DB == nil || m.Now == nil {
		return TagUpdateResult{}, errors.New("metadata service misconfigured")
	}

	normalized, err := key.Normalize(rawKey)
	if err != nil {
		return TagUpdateResult{}, err
	}

	meta, err := storage.GetMetadata(ctx, m.DB, normalized)
	if err != nil {
		return TagUpdateResult{}, err
	}

	before := parseTags(meta.Tags)
	updated := MergeTags(meta.Tags, add, remove)
	after := parseTags(updated)

	added := diff(after, before)
	removed := diff(before, after)

	if updated != meta.Tags {
		meta.Tags = updated
		meta.Modified = m.Now().UTC()
		if err := storage.UpdateMetadata(ctx, m.DB, meta); err != nil {
			return TagUpdateResult{}, err
		}
	}

	return TagUpdateResult{
		Metadata: meta,
		Added:    added,
		Removed:  removed,
	}, nil
}

func diff(a, b []string) []string {
	if len(a) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(b))
	for _, tag := range b {
		set[tag] = struct{}{}
	}
	var out []string
	for _, tag := range a {
		if _, ok := set[tag]; ok {
			continue
		}
		out = append(out, tag)
	}
	return out
}

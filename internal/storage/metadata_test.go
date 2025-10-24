package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/model"
)

func TestInsertAndGetMetadata(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), ".meta.db")

	db, err := InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	now := time.Unix(1_700_000_000, 0)
	meta := model.Metadata{
		Key:         "go/foo",
		Type:        "text",
		Created:     now,
		Modified:    now,
		Description: "desc",
		Tags:        "tag1,tag2",
	}
	meta.Created = meta.Created.UTC()
	meta.Modified = meta.Modified.UTC()

	if err := InsertMetadata(ctx, db, meta); err != nil {
		t.Fatalf("InsertMetadata error = %v", err)
	}

	got, err := GetMetadata(ctx, db, "go/foo")
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}

	if got != meta {
		t.Fatalf("GetMetadata = %+v, want %+v", got, meta)
	}
}

func TestInsertMetadataDuplicate(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "meta.db")

	db, err := InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	now := time.Unix(1_700_000_000, 0)
	meta := model.Metadata{
		Key:      "dup/key",
		Type:     "text",
		Created:  now,
		Modified: now,
	}

	if err := InsertMetadata(ctx, db, meta); err != nil {
		t.Fatalf("first insert error = %v", err)
	}

	if err := InsertMetadata(ctx, db, meta); err != ErrMetadataDuplicate {
		t.Fatalf("second insert error = %v, want ErrMetadataDuplicate", err)
	}
}

func TestGetMetadataNotFound(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "meta.db")

	db, err := InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if _, err := GetMetadata(ctx, db, "missing"); err != ErrMetadataNotFound {
		t.Fatalf("expected ErrMetadataNotFound, got %v", err)
	}
}

func TestListMetadataOrdersByCreatedDesc(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "meta.db")

	db, err := InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	base := time.Unix(1_700_000_000, 0)
	entries := []model.Metadata{
		{Key: "first", Type: "text", Created: base.Add(-2 * time.Hour), Modified: base.Add(-2 * time.Hour)},
		{Key: "second", Type: "text", Created: base.Add(-1 * time.Hour), Modified: base.Add(-1 * time.Hour)},
		{Key: "third", Type: "url", Created: base, Modified: base},
	}

	for _, meta := range entries {
		if err := InsertMetadata(ctx, db, meta); err != nil {
			t.Fatalf("InsertMetadata %q error = %v", meta.Key, err)
		}
	}

	list, err := ListMetadata(ctx, db)
	if err != nil {
		t.Fatalf("ListMetadata error = %v", err)
	}

	if len(list) != len(entries) {
		t.Fatalf("ListMetadata len = %d, want %d", len(list), len(entries))
	}

	wantOrder := []string{"third", "second", "first"}
	for i, key := range wantOrder {
		if list[i].Key != key {
			t.Fatalf("ListMetadata[%d].Key = %q, want %q", i, list[i].Key, key)
		}
	}
}

func TestDeleteMetadata(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "meta.db")

	db, err := InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	now := time.Unix(1_700_000_000, 0)
	meta := model.Metadata{
		Key:      "delete/me",
		Type:     "text",
		Created:  now,
		Modified: now,
	}

	if err := InsertMetadata(ctx, db, meta); err != nil {
		t.Fatalf("InsertMetadata error = %v", err)
	}

	if err := DeleteMetadata(ctx, db, meta.Key); err != nil {
		t.Fatalf("DeleteMetadata error = %v", err)
	}

	if _, err := GetMetadata(ctx, db, meta.Key); err != ErrMetadataNotFound {
		t.Fatalf("expected ErrMetadataNotFound after delete, got %v", err)
	}

	if err := DeleteMetadata(ctx, db, meta.Key); err != ErrMetadataNotFound {
		t.Fatalf("second delete expected ErrMetadataNotFound, got %v", err)
	}
}

func TestUpdateMetadata(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "meta.db")

	db, err := InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	created := time.Unix(1_700_000_000, 0).UTC()
	initial := model.Metadata{
		Key:         "go/foo",
		Type:        "text",
		Created:     created,
		Modified:    created,
		Description: "original",
		Tags:        "first,second",
	}

	if err := InsertMetadata(ctx, db, initial); err != nil {
		t.Fatalf("InsertMetadata error = %v", err)
	}

	updated := initial
	updated.Type = "url"
	updated.Modified = created.Add(time.Hour).UTC()
	updated.Description = "updated description"
	updated.Tags = "updated"

	if err := UpdateMetadata(ctx, db, updated); err != nil {
		t.Fatalf("UpdateMetadata error = %v", err)
	}

	got, err := GetMetadata(ctx, db, updated.Key)
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}

	if got.Type != updated.Type {
		t.Fatalf("Type = %q, want %q", got.Type, updated.Type)
	}
	if !got.Modified.Equal(updated.Modified) {
		t.Fatalf("Modified = %v, want %v", got.Modified, updated.Modified)
	}
	if got.Description != updated.Description {
		t.Fatalf("Description = %q, want %q", got.Description, updated.Description)
	}
	if got.Tags != updated.Tags {
		t.Fatalf("Tags = %q, want %q", got.Tags, updated.Tags)
	}
	if !got.Created.Equal(created) {
		t.Fatalf("Created changed = %v, want %v", got.Created, created)
	}
}

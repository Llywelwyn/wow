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
		Key:         "go/foo",
		Type:        "text",
		Created:     now,
		Modified:    now,
		Description: "desc",
		Tags:        "tag1,tag2",
	}

	if err := InsertMetadata(ctx, db, meta); err != nil {
		t.Fatalf("InsertMetadata error = %v", err)
	}

	got, err := GetMetadata(ctx, db, "go/foo")
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}

	if got.Key != meta.Key || got.Type != meta.Type {
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

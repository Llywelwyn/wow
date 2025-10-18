package services

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/storage"
)

func TestRemoverRemoveSuccess(t *testing.T) {
	base := t.TempDir()
	dbPath := filepath.Join(base, "meta.db")
	db, err := storage.InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	saver := &Saver{
		BaseDir: base,
		DB:      db,
		Now: func() time.Time {
			return time.Unix(1_700_000_000, 0)
		},
	}

	ctx := context.Background()
	if _, err := saver.Save(ctx, SaveRequest{Key: "go/foo", Reader: strings.NewReader("content")}); err != nil {
		t.Fatalf("Save error = %v", err)
	}

	remover := &Remover{BaseDir: base, DB: db}
	if err := remover.Remove(ctx, "go/foo"); err != nil {
		t.Fatalf("Remove error = %v", err)
	}

	if _, err := storage.GetMetadata(ctx, db, "go/foo"); err != storage.ErrMetadataNotFound {
		t.Fatalf("expected metadata deleted, got %v", err)
	}

	path := filepath.Join(base, "go", "foo")
	if exists, err := storage.Exists(path); err != nil {
		t.Fatalf("Exists error = %v", err)
	} else if exists {
		t.Fatalf("expected file removed")
	}
}

func TestRemoverMissingKey(t *testing.T) {
	base := t.TempDir()
	dbPath := filepath.Join(base, "meta.db")
	db, err := storage.InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	remover := &Remover{BaseDir: base, DB: db}
	err = remover.Remove(context.Background(), "missing")
	if err != storage.ErrMetadataNotFound {
		t.Fatalf("expected ErrMetadataNotFound, got %v", err)
	}
}

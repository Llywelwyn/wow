package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/model"
)

func TestInsertAndGetSnippet(t *testing.T) {
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
	sn := model.Snippet{
		Key:         "go/foo",
		Type:        "text",
		Created:     now,
		Modified:    now,
		Description: "desc",
		Tags:        "tag1,tag2",
	}

	if err := InsertSnippet(ctx, db, sn); err != nil {
		t.Fatalf("InsertSnippet error = %v", err)
	}

	got, err := GetSnippet(ctx, db, "go/foo")
	if err != nil {
		t.Fatalf("GetSnippet error = %v", err)
	}

	if got.Key != sn.Key || got.Type != sn.Type {
		t.Fatalf("GetSnippet = %+v, want %+v", got, sn)
	}
}

func TestInsertSnippetDuplicate(t *testing.T) {
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
	sn := model.Snippet{
		Key:      "dup/key",
		Type:     "text",
		Created:  now,
		Modified: now,
	}

	if err := InsertSnippet(ctx, db, sn); err != nil {
		t.Fatalf("first insert error = %v", err)
	}

	if err := InsertSnippet(ctx, db, sn); err != ErrSnippetDuplicate {
		t.Fatalf("second insert error = %v, want ErrSnippetDuplicate", err)
	}
}

func TestGetSnippetNotFound(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "meta.db")

	db, err := InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if _, err := GetSnippet(ctx, db, "missing"); err != ErrSnippetNotFound {
		t.Fatalf("expected ErrSnippetNotFound, got %v", err)
	}
}

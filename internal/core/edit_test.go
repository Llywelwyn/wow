package core

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/storage"
)

func newEditEnv(t *testing.T) (*Editor, *Saver, context.Context) {
	t.Helper()

	base := t.TempDir()
	dbPath := filepath.Join(base, "meta.db")
	db, err := storage.InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()

	saver := &Saver{
		BaseDir: base,
		DB:      db,
		Now: func() time.Time {
			return time.Unix(1_700_000_000, 0).UTC()
		},
	}

	editor := &Editor{
		BaseDir: base,
		DB:      db,
		Now: func() time.Time {
			return time.Unix(1_700_000_100, 0).UTC()
		},
	}

	return editor, saver, ctx
}

func TestEditorEditMisconfigured(t *testing.T) {
	editor := &Editor{}
	if _, err := editor.Edit(context.Background(), "go/foo"); err == nil {
		t.Fatalf("expected misconfigured error")
	}
}

func TestEditorEditInvalidKey(t *testing.T) {
	editor := &Editor{
		DB:  &sql.DB{},
		Now: time.Now,
		Open: func(ctx context.Context, path string) error {
			return nil
		},
	}

	if _, err := editor.Edit(context.Background(), "bad key"); err == nil {
		t.Fatalf("expected invalid key error")
	}
}

func TestEditorEditMetadataNotFound(t *testing.T) {
	editor, _, ctx := newEditEnv(t)
	editor.Open = func(ctx context.Context, path string) error { return nil }

	if _, err := editor.Edit(ctx, "missing/key"); err != storage.ErrMetadataNotFound {
		t.Fatalf("expected ErrMetadataNotFound, got %v", err)
	}
}

func TestEditorEditOpenError(t *testing.T) {
	editor, saver, ctx := newEditEnv(t)

	if _, err := saver.Save(ctx, SaveRequest{
		Key:    "go/foo",
		Reader: strings.NewReader("package main\n"),
	}); err != nil {
		t.Fatalf("Save error = %v", err)
	}

	editor.Open = func(ctx context.Context, path string) error {
		return errors.New("editor failed")
	}

	if _, err := editor.Edit(ctx, "go/foo"); err == nil {
		t.Fatalf("expected editor failure error")
	}
}

func TestEditorEditUpdatesMetadataWhenFileChanges(t *testing.T) {
	editor, saver, ctx := newEditEnv(t)

	if _, err := saver.Save(ctx, SaveRequest{
		Key:    "go/foo",
		Reader: strings.NewReader("package main\n"),
	}); err != nil {
		t.Fatalf("Save error = %v", err)
	}

	editor.Open = func(ctx context.Context, path string) error {
		return os.WriteFile(path, []byte("https://example.com\n"), 0o600)
	}

	meta, err := editor.Edit(ctx, "go/foo")
	if err != nil {
		t.Fatalf("Edit error = %v", err)
	}

	wantModified := editor.Now()
	if !meta.Modified.Equal(wantModified) {
		t.Fatalf("Modified = %v, want %v", meta.Modified, wantModified)
	}

	stored, err := storage.GetMetadata(ctx, editor.DB, "go/foo")
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}

	if !stored.Modified.Equal(wantModified) {
		t.Fatalf("stored Modified = %v, want %v", stored.Modified, wantModified)
	}
}

func TestEditorEditNoChangeLeavesMetadataUntouched(t *testing.T) {
	editor, saver, ctx := newEditEnv(t)

	if _, err := saver.Save(ctx, SaveRequest{
		Key:    "go/foo",
		Reader: strings.NewReader("package main\n"),
	}); err != nil {
		t.Fatalf("Save error = %v", err)
	}

	original, err := storage.GetMetadata(ctx, editor.DB, "go/foo")
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}

	editor.Open = func(ctx context.Context, path string) error {
		return nil
	}

	meta, err := editor.Edit(ctx, "go/foo")
	if err != nil {
		t.Fatalf("Edit error = %v", err)
	}

	if meta != original {
		t.Fatalf("Metadata changed = %q, want %q", meta.Type, original.Type)
	}
}

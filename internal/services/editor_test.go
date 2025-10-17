package services

import (
	"context"
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

func TestEditorEditUpdatesMetadataWhenFileChanges(t *testing.T) {
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
		return os.WriteFile(path, []byte("https://example.com\n"), 0o600)
	}

	meta, err := editor.Edit(ctx, "go/foo")
	if err != nil {
		t.Fatalf("Edit error = %v", err)
	}

	if meta.Type != "url" {
		t.Fatalf("Type = %q, want url", meta.Type)
	}

	wantModified := editor.Now().UTC()
	if !meta.Modified.Equal(wantModified) {
		t.Fatalf("Modified = %v, want %v", meta.Modified, wantModified)
	}

	stored, err := storage.GetMetadata(ctx, editor.DB, "go/foo")
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}

	if stored.Type != "url" {
		t.Fatalf("stored Type = %q, want url", stored.Type)
	}
	if !stored.Modified.Equal(wantModified) {
		t.Fatalf("stored Modified = %v, want %v", stored.Modified, wantModified)
	}
	if !stored.Created.Equal(original.Created) {
		t.Fatalf("stored Created changed = %v, want %v", stored.Created, original.Created)
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

	if meta.Modified != original.Modified {
		t.Fatalf("Modified changed = %v, want %v", meta.Modified, original.Modified)
	}
	if meta.Type != original.Type {
		t.Fatalf("Type changed = %q, want %q", meta.Type, original.Type)
	}
}

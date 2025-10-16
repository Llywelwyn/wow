package core

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/storage"
)

func newTestSaver(t *testing.T) (*Saver, context.Context) {
	t.Helper()

	base := t.TempDir()
	dbPath := filepath.Join(base, "meta.db")
	db, err := storage.InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	s := &Saver{
		BaseDir: base,
		DB:      db,
		Now: func() time.Time {
			return time.Unix(1_700_000_000, 0)
		},
	}

	return s, context.Background()
}

func TestSaverSaveExplicitKey(t *testing.T) {
	s, ctx := newTestSaver(t)

	_, err := s.Save(ctx, SaveRequest{
		Key:    "go/foo",
		Reader: strings.NewReader("package main\n"),
		Tags:   []string{"Go", "utils", "go"},
	})
	if err != nil {
		t.Fatalf("Save error = %v", err)
	}

	path, err := filepath.Abs(filepath.Join(s.BaseDir, "go", "foo"))
	if err != nil {
		t.Fatalf("Resolve path error = %v", err)
	}

	if exists, err := storage.Exists(path); err != nil || !exists {
		t.Fatalf("snippet file should exist, err=%v exists=%v", err, exists)
	}

	meta, err := storage.GetMetadata(ctx, s.DB, "go/foo")
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}
	if meta.Type != "text" {
		t.Fatalf("metadata type = %q, want text", meta.Type)
	}
	if meta.Tags != "go,utils" {
		t.Fatalf("tags = %q, want go,utils", meta.Tags)
	}
}

func TestSaverAutoKey(t *testing.T) {
	s, ctx := newTestSaver(t)

	result, err := s.Save(ctx, SaveRequest{
		Reader: strings.NewReader("https://example.com"),
	})
	if err != nil {
		t.Fatalf("Save error = %v", err)
	}

	wantKey := "auto/1700000000"
	if result.Key != wantKey {
		t.Fatalf("auto key = %q, want %q", result.Key, wantKey)
	}

	meta, err := storage.GetMetadata(ctx, s.DB, wantKey)
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}
	if meta.Type != "url" {
		t.Fatalf("metadata type = %q, want url", meta.Type)
	}
}

func TestSaverAutoKeyWithCollision(t *testing.T) {
	s, ctx := newTestSaver(t)

	// Seed an existing file to force collision.
	path := filepath.Join(s.BaseDir, "auto", "1700000000")
	if err := storage.Save(path, strings.NewReader("existing")); err != nil {
		t.Fatalf("pre-save error = %v", err)
	}

	result, err := s.Save(ctx, SaveRequest{
		Reader: strings.NewReader("new snippet"),
	})
	if err != nil {
		t.Fatalf("Save error = %v", err)
	}

	if result.Key != "auto/1700000000-1" {
		t.Fatalf("auto key = %q, want auto/1700000000-1", result.Key)
	}
}

func TestSaverDuplicateKey(t *testing.T) {
	s, ctx := newTestSaver(t)

	if _, err := s.Save(ctx, SaveRequest{
		Key:    "dup/foo",
		Reader: strings.NewReader("snippet"),
	}); err != nil {
		t.Fatalf("first Save error = %v", err)
	}

	if _, err := s.Save(ctx, SaveRequest{
		Key:    "dup/foo",
		Reader: strings.NewReader("snippet"),
	}); err != ErrSnippetExists {
		t.Fatalf("second Save error = %v, want ErrSnippetExists", err)
	}
}

func TestSaverEmptyInput(t *testing.T) {
	s, ctx := newTestSaver(t)

	_, err := s.Save(ctx, SaveRequest{
		Key:    "empty/foo",
		Reader: strings.NewReader(""),
	})
	if err == nil {
		t.Fatalf("expected error for empty content")
	}
}

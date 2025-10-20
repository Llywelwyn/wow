package command

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/storage"
)

func newRemoveCommandEnv(t *testing.T) (*RemoveCommand, *SaveCommand, func()) {
	t.Helper()

	base := t.TempDir()
	dbPath := filepath.Join(base, "meta.db")
	db, err := storage.InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}

	saveCmd := &SaveCommand{
		BaseDir: base,
		DB:      db,
		Now: func() time.Time {
			return time.Unix(1_700_000_000, 0)
		},
	}

	cmd := &RemoveCommand{BaseDir: base, DB: db}

	cleanup := func() {
		_ = db.Close()
	}

	return cmd, saveCmd, cleanup
}

func TestRemoveCommandDeletesSnippet(t *testing.T) {
	cmd, saveCmd, cleanup := newRemoveCommandEnv(t)
	defer cleanup()

	ctx := context.Background()
	saveCmd.Input = strings.NewReader("content")
	saveCmd.Output = bytes.NewBuffer(nil)
	if err := saveCmd.Execute([]string{"go/foo"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	if err := cmd.Execute([]string{"go/foo"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	if _, err := storage.GetMetadata(ctx, cmd.DB, "go/foo"); !errors.Is(err, storage.ErrMetadataNotFound) {
		t.Fatalf("expected ErrMetadataNotFound, got %v", err)
	}

	path := filepath.Join(cmd.BaseDir, "go", "foo")
	if exists, err := storage.Exists(path); err != nil {
		t.Fatalf("Exists error = %v", err)
	} else if exists {
		t.Fatalf("expected file removed")
	}
}

func TestRemoveCommandRequiresKey(t *testing.T) {
	cmd, _, cleanup := newRemoveCommandEnv(t)
	defer cleanup()

	if err := cmd.Execute(nil); err == nil || err.Error() != "key required" {
		t.Fatalf("expected key required error, got %v", err)
	}
}

func TestRemoveCommandMissingSnippet(t *testing.T) {
	cmd, _, cleanup := newRemoveCommandEnv(t)
	defer cleanup()

	err := cmd.Execute([]string{"missing"})
	if !errors.Is(err, storage.ErrMetadataNotFound) {
		t.Fatalf("expected ErrMetadataNotFound, got %v", err)
	}
}

package command

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/storage"
)

type editStub struct {
	called bool
	modify func(path string) error
	err    error
}

func (s *editStub) Call(ctx context.Context, path string) error {
	s.called = true
	if ctx == nil {
		return errors.New("missing context")
	}
	if s.err != nil {
		return s.err
	}
	if s.modify != nil {
		return s.modify(path)
	}
	return nil
}

func setupEditCommand(t *testing.T) (*EditCommand, *SaveCommand, *editStub, func()) {
	t.Helper()

	base := t.TempDir()
	dbPath := filepath.Join(base, "meta.db")
	db, err := storage.InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}

	stub := &editStub{}

	cmd := &EditCommand{
		BaseDir: base,
		DB:      db,
		Now: func() time.Time {
			return time.Unix(1_700_000_100, 0)
		},
		Open: stub.Call,
	}

	saveCmd := &SaveCommand{
		BaseDir: base,
		DB:      db,
		Now: func() time.Time {
			return time.Unix(1_700_000_000, 0)
		},
	}

	cleanup := func() { _ = db.Close() }
	return cmd, saveCmd, stub, cleanup
}

func TestEditCommandName(t *testing.T) {
	cmd := &EditCommand{}
	if cmd.Name() != "edit" {
		t.Fatalf("Name = %q, want edit", cmd.Name())
	}
}

func TestEditCommandRequiresConfiguration(t *testing.T) {
	cmd := &EditCommand{}
	if err := cmd.Execute([]string{"key"}); err == nil {
		t.Fatalf("expected configuration error")
	}
}

func TestEditCommandRequiresSingleKey(t *testing.T) {
	cmd, _, _, cleanup := setupEditCommand(t)
	defer cleanup()

	if err := cmd.Execute(nil); err == nil {
		t.Fatalf("expected error for missing key")
	}

	if err := cmd.Execute([]string{"one", "two"}); err == nil {
		t.Fatalf("expected error for too many args")
	}
}

func TestEditCommandNoChangesLeavesMetadata(t *testing.T) {
	cmd, saveCmd, stub, cleanup := setupEditCommand(t)
	defer cleanup()

	saveCmd.Input = bytes.NewBufferString("content")
	saveCmd.Output = bytes.NewBuffer(nil)
	if err := saveCmd.Execute([]string{"go/foo"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	if err := cmd.Execute([]string{"go/foo"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if !stub.called {
		t.Fatalf("expected stub to be called")
	}

	meta, err := storage.GetMetadata(context.Background(), cmd.DB, "go/foo")
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}
	if got, want := meta.Modified.UTC(), time.Unix(1_700_000_000, 0).UTC(); !got.Equal(want) {
		t.Fatalf("Modified = %v, want %v", got, want)
	}
}

func TestEditCommandUpdatesMetadataOnChange(t *testing.T) {
	cmd, saveCmd, stub, cleanup := setupEditCommand(t)
	defer cleanup()

	saveCmd.Input = bytes.NewBufferString("content")
	saveCmd.Output = bytes.NewBuffer(nil)
	if err := saveCmd.Execute([]string{"urls/site"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	stub.modify = func(path string) error {
		return os.WriteFile(path, []byte("https://example.com"), 0o600)
	}

	if err := cmd.Execute([]string{"urls/site"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	meta, err := storage.GetMetadata(context.Background(), cmd.DB, "urls/site")
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}
	if meta.Type != "url" {
		t.Fatalf("Type = %q, want url", meta.Type)
	}
	wantTime := time.Unix(1_700_000_100, 0).UTC()
	if got := meta.Modified.UTC(); !got.Equal(wantTime) {
		t.Fatalf("Modified = %v, want %v", got, wantTime)
	}
}

func TestEditCommandPropagatesErrors(t *testing.T) {
	cmd, saveCmd, stub, cleanup := setupEditCommand(t)
	defer cleanup()

	saveCmd.Input = bytes.NewBufferString("content")
	saveCmd.Output = bytes.NewBuffer(nil)
	if err := saveCmd.Execute([]string{"go/bar"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	stub.err = errors.New("boom")
	if err := cmd.Execute([]string{"go/bar"}); err != stub.err {
		t.Fatalf("expected %v, got %v", stub.err, err)
	}
}

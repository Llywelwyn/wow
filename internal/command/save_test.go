package command

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/services"
	"github.com/llywelwyn/wow/internal/storage"
)

func newTestSaveCommand(t *testing.T) (*SaveCommand, func()) {
	t.Helper()

	base := t.TempDir()
	dbPath := filepath.Join(base, "meta.db")
	db, err := storage.InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}

	saver := &services.Saver{
		BaseDir: base,
		DB:      db,
		Now: func() time.Time {
			return time.Unix(1_700_000_000, 0)
		},
	}

	cmd := &SaveCommand{
		Saver: saver,
		// Input and Output assigned per test.
	}

	cleanup := func() {
		_ = db.Close()
	}
	return cmd, cleanup
}

func TestSaveCommandAutoKey(t *testing.T) {
	cmd, cleanup := newTestSaveCommand(t)
	defer cleanup()

	var out bytes.Buffer
	cmd.Input = strings.NewReader("foo bar")
	cmd.Output = &out

	if err := cmd.Execute(nil); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	got := strings.TrimSpace(out.String())
	if got != "auto/1700000000" {
		t.Fatalf("output key = %q, want auto/1700000000", got)
	}
}

func TestSaveCommandExplicitKey(t *testing.T) {
	cmd, cleanup := newTestSaveCommand(t)
	defer cleanup()

	var out bytes.Buffer
	cmd.Input = strings.NewReader("foo bar")
	cmd.Output = &out

	if err := cmd.Execute([]string{"go/foo"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	got := strings.TrimSpace(out.String())
	if got != "go/foo" {
		t.Fatalf("output key = %q, want go/foo", got)
	}
}

func TestSaveCommandTagsFlag(t *testing.T) {
	cmd, cleanup := newTestSaveCommand(t)
	defer cleanup()

	var out bytes.Buffer
	cmd.Input = strings.NewReader("content")
	cmd.Output = &out

	if err := cmd.Execute([]string{"--tag", "foo,Bar"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	got := strings.TrimSpace(out.String())
	if got != "auto/1700000000" {
		t.Fatalf("output key = %q, want auto/1700000000", got)
	}

	meta, err := storage.GetMetadata(context.Background(), cmd.Saver.DB, got)
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}
	if meta.Tags != "foo,bar" {
		t.Fatalf("tags stored = %q, want foo,bar", meta.Tags)
	}
}

func TestSaveCommandImplicitTags(t *testing.T) {
	cmd, cleanup := newTestSaveCommand(t)
	defer cleanup()

	var out bytes.Buffer
	cmd.Input = strings.NewReader("content")
	cmd.Output = &out

	if err := cmd.Execute([]string{"go/foo", "@Foo", "@bar"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	meta, err := storage.GetMetadata(context.Background(), cmd.Saver.DB, "go/foo")
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}
	if meta.Tags != "foo,bar" {
		t.Fatalf("tags stored = %q, want foo,bar", meta.Tags)
	}
}

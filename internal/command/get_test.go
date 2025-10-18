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

func setupGetTest(t *testing.T) (Config, *services.Saver, func()) {
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

	cfg := Config{
		BaseDir: base,
		DB:      db,
		Clock: func() time.Time {
			return time.Unix(1_700_000_100, 0)
		},
	}

	cleanup := func() { _ = db.Close() }
	return cfg, saver, cleanup
}

func TestGetCommandReadsSnippet(t *testing.T) {
	cfg, saver, cleanup := setupGetTest(t)
	defer cleanup()

	saveCmd := &SaveCommand{
		Saver:  saver,
		Input:  strings.NewReader("hello world"),
		Output: bytes.NewBuffer(nil),
	}
	if err := saveCmd.Execute([]string{"go/foo"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	var out bytes.Buffer
	cfg.Output = &out
	getCmd := NewGetCommand(cfg)

	if err := getCmd.Execute([]string{"go/foo"}); err != nil {
		t.Fatalf("Get Execute error = %v", err)
	}

	if out.String() != "hello world" {
		t.Fatalf("output = %q, want %q", out.String(), "hello world")
	}
}

func TestGetCommandImplicitAddTag(t *testing.T) {
	cfg, saver, cleanup := setupGetTest(t)
	defer cleanup()

	saveCmd := &SaveCommand{
		Saver:  saver,
		Input:  strings.NewReader("content"),
		Output: bytes.NewBuffer(nil),
	}
	if err := saveCmd.Execute([]string{"go/foo"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	var out bytes.Buffer
	cfg.Output = &out
	getCmd := NewGetCommand(cfg)

	if err := getCmd.Execute([]string{"go/foo", "@bar"}); err != nil {
		t.Fatalf("Get Execute error = %v", err)
	}

	meta, err := storage.GetMetadata(context.Background(), cfg.DB, "go/foo")
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}
	if meta.Tags != "bar" {
		t.Fatalf("tags = %q, want bar", meta.Tags)
	}
}

func TestGetCommandRemoveTag(t *testing.T) {
	cfg, saver, cleanup := setupGetTest(t)
	defer cleanup()

	saveCmd := &SaveCommand{
		Saver:  saver,
		Input:  strings.NewReader("content"),
		Output: bytes.NewBuffer(nil),
	}
	if err := saveCmd.Execute([]string{"go/foo", "@foo"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	var out bytes.Buffer
	cfg.Output = &out
	getCmd := NewGetCommand(cfg)

	if err := getCmd.Execute([]string{"go/foo", "-@foo"}); err != nil {
		t.Fatalf("Get Execute error = %v", err)
	}

	meta, err := storage.GetMetadata(context.Background(), cfg.DB, "go/foo")
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}
	if meta.Tags != "" {
		t.Fatalf("tags = %q, want empty", meta.Tags)
	}
}

func TestGetCommandFlagTagging(t *testing.T) {
	cfg, saver, cleanup := setupGetTest(t)
	defer cleanup()

	saveCmd := &SaveCommand{
		Saver:  saver,
		Input:  strings.NewReader("content"),
		Output: bytes.NewBuffer(nil),
	}
	if err := saveCmd.Execute([]string{"go/foo", "@foo"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	var out bytes.Buffer
	cfg.Output = &out
	getCmd := NewGetCommand(cfg)

	if err := getCmd.Execute([]string{"go/foo", "--tag", "bar", "--untag", "foo"}); err != nil {
		t.Fatalf("Get Execute error = %v", err)
	}

	meta, err := storage.GetMetadata(context.Background(), cfg.DB, "go/foo")
	if err != nil {
		t.Fatalf("GetMetadata error = %v", err)
	}
	if meta.Tags != "bar" {
		t.Fatalf("tags = %q, want bar", meta.Tags)
	}
}

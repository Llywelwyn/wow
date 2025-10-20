package command

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/storage"
)

type fnSpy struct {
	called bool
	target string
	err    error
}

func (s *fnSpy) Call(ctx context.Context, target string) error {
	s.called = true
	s.target = target
	if ctx == nil {
		return errors.New("missing context")
	}
	return s.err
}

func setupOpenCommand(t *testing.T) (*OpenCommand, *SaveCommand, *fnSpy, *fnSpy, func()) {
	t.Helper()

	base := t.TempDir()
	dbPath := filepath.Join(base, "meta.db")
	db, err := storage.InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}

	openSpy := &fnSpy{}
	pagerSpy := &fnSpy{}

	cmd := &OpenCommand{
		BaseDir:   base,
		DB:        db,
		OpenFunc:  openSpy.Call,
		PagerFunc: pagerSpy.Call,
	}

	saveCmd := &SaveCommand{
		BaseDir: base,
		DB:      db,
		Now: func() time.Time {
			return time.Unix(1_700_000_000, 0)
		},
	}

	cleanup := func() { _ = db.Close() }
	return cmd, saveCmd, openSpy, pagerSpy, cleanup
}

func TestOpenCommandName(t *testing.T) {
	cmd := &OpenCommand{}
	if cmd.Name() != "open" {
		t.Fatalf("Name = %q, want open", cmd.Name())
	}
}

func TestOpenCommandRequiresConfiguration(t *testing.T) {
	cmd := &OpenCommand{}
	if err := cmd.Execute([]string{"key"}); err == nil {
		t.Fatalf("expected configuration error")
	}
}

func TestOpenCommandRequiresKey(t *testing.T) {
	cmd, _, _, _, cleanup := setupOpenCommand(t)
	defer cleanup()

	if err := cmd.Execute(nil); err == nil {
		t.Fatalf("expected error for missing key")
	}
}

func TestOpenCommandOpensFilePath(t *testing.T) {
	cmd, saveCmd, openSpy, pagerSpy, cleanup := setupOpenCommand(t)
	defer cleanup()

	saveCmd.Input = bytes.NewBufferString("hello")
	saveCmd.Output = bytes.NewBuffer(nil)
	if err := saveCmd.Execute([]string{"go/foo"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	if err := cmd.Execute([]string{"go/foo"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if !openSpy.called {
		t.Fatalf("expected open spy to be called")
	}
	if openSpy.target != filepath.Join(cmd.BaseDir, "go", "foo") {
		t.Fatalf("open target = %q", openSpy.target)
	}
	if pagerSpy.called {
		t.Fatalf("did not expect pager call")
	}
}

func TestOpenCommandPagerFlagUsesPager(t *testing.T) {
	cmd, saveCmd, openSpy, pagerSpy, cleanup := setupOpenCommand(t)
	defer cleanup()

	saveCmd.Input = bytes.NewBufferString("content")
	saveCmd.Output = bytes.NewBuffer(nil)
	if err := saveCmd.Execute([]string{"text/snippet"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	if err := cmd.Execute([]string{"--pager", "text/snippet"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if !pagerSpy.called {
		t.Fatalf("expected pager spy to be called")
	}
	if pagerSpy.target != filepath.Join(cmd.BaseDir, "text", "snippet") {
		t.Fatalf("pager target = %q", pagerSpy.target)
	}
	if openSpy.called {
		t.Fatalf("did not expect open spy call")
	}
}

func TestOpenCommandOpensURLs(t *testing.T) {
	cmd, saveCmd, openSpy, _, cleanup := setupOpenCommand(t)
	defer cleanup()

	saveCmd.Input = bytes.NewBufferString("https://example.com\n")
	saveCmd.Output = bytes.NewBuffer(nil)
	if err := saveCmd.Execute([]string{"urls/site"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	if err := cmd.Execute([]string{"urls/site"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if openSpy.target != "https://example.com" {
		t.Fatalf("open target = %q", openSpy.target)
	}
}

func TestOpenCommandPropagatesErrors(t *testing.T) {
	cmd, saveCmd, openSpy, _, cleanup := setupOpenCommand(t)
	defer cleanup()

	openSpy.err = errors.New("boom")

	saveCmd.Input = bytes.NewBufferString("content")
	saveCmd.Output = bytes.NewBuffer(nil)
	if err := saveCmd.Execute([]string{"go/bar"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	if err := cmd.Execute([]string{"go/bar"}); err != openSpy.err {
		t.Fatalf("expected %v, got %v", openSpy.err, err)
	}
}

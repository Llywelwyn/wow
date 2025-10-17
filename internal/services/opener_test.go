package services

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/storage"
)

type stubRunner struct {
	calledWith []string
	err        error
}

func (s *stubRunner) run(_ context.Context, target string) error {
	s.calledWith = append(s.calledWith, target)
	return s.err
}

func newOpenerEnv(t *testing.T) (*Opener, context.Context, func(string, string) error) {
	t.Helper()

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

	save := func(key, content string) error {
		_, err := saver.Save(ctx, SaveRequest{
			Key:    key,
			Reader: strings.NewReader(content),
		})
		return err
	}

	opener := &Opener{
		BaseDir: base,
		DB:      db,
	}

	return opener, ctx, save
}

func setupOpener(t *testing.T) (*Opener, context.Context, func(string, string) error, *stubRunner, *stubRunner) {
	opener, ctx, save := newOpenerEnv(t)

	open := &stubRunner{}
	pager := &stubRunner{}

	opener.OpenFunc = open.run
	opener.PagerFunc = pager.run

	return opener, ctx, save, open, pager
}

func TestOpenerOpenTextCallsOpenerWithPath(t *testing.T) {
	opener, ctx, save, open, pager := setupOpener(t)

	if err := save("go/text", "hello world"); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	if err := opener.Open(ctx, "go/text", OpenOptions{}); err != nil {
		t.Fatalf("Open error = %v", err)
	}

	if len(open.calledWith) != 1 {
		t.Fatalf("opener called %d times, want 1", len(open.calledWith))
	}
	if !strings.HasSuffix(open.calledWith[0], filepath.Join("go", "text")) {
		t.Fatalf("opener target = %q, want path", open.calledWith[0])
	}

	if len(pager.calledWith) != 0 {
		t.Fatalf("pager should not be called")
	}
}

func TestOpenerOpenURLUsesFirstLine(t *testing.T) {
	opener, ctx, save, open, _ := setupOpener(t)

	if err := save("urls/git", "https://example.com\nsecond line"); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	if err := opener.Open(ctx, "urls/git", OpenOptions{}); err != nil {
		t.Fatalf("Open error = %v", err)
	}

	if len(open.calledWith) != 1 {
		t.Fatalf("opener called %d times, want 1", len(open.calledWith))
	}
	if open.calledWith[0] != "https://example.com" {
		t.Fatalf("opener target = %q, want url", open.calledWith[0])
	}
}

func TestOpenerOpenUrlFallsBackToPathWhenEmpty(t *testing.T) {
	opener, ctx, save, open, _ := setupOpener(t)

	if err := save("urls/empty", "\n\n"); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	if err := opener.Open(ctx, "urls/empty", OpenOptions{}); err != nil {
		t.Fatalf("Open error = %v", err)
	}

	if len(open.calledWith) != 1 {
		t.Fatalf("opener called %d times, want 1", len(open.calledWith))
	}
	if !strings.HasSuffix(open.calledWith[0], filepath.Join("urls", "empty")) {
		t.Fatalf("opener target = %q, want file path", open.calledWith[0])
	}
}

func TestOpenerUsePager(t *testing.T) {
	opener, ctx, save, open, pager := setupOpener(t)

	if err := save("go/text", "hello"); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	if err := opener.Open(ctx, "go/text", OpenOptions{UsePager: true}); err != nil {
		t.Fatalf("Open error = %v", err)
	}

	if len(pager.calledWith) != 1 {
		t.Fatalf("pager called %d times, want 1", len(pager.calledWith))
	}
	if len(open.calledWith) != 0 {
		t.Fatalf("opener should not be called when pager requested")
	}
}

func TestOpenerMissingMetadata(t *testing.T) {
	opener, ctx, _, open, pager := setupOpener(t)

	if err := opener.Open(ctx, "missing/key", OpenOptions{}); !errors.Is(err, storage.ErrMetadataNotFound) {
		t.Fatalf("expected ErrMetadataNotFound, got %v", err)
	}
	if len(open.calledWith) != 0 || len(pager.calledWith) != 0 {
		t.Fatalf("no runners should be called on error")
	}
}

func TestOpenerPropagatesRunnerErrors(t *testing.T) {
	opener, ctx, save, open, pager := setupOpener(t)

	open.err = errors.New("boom")

	if err := save("go/text", "hello"); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	if err := opener.Open(ctx, "go/text", OpenOptions{}); !errors.Is(err, open.err) {
		t.Fatalf("expected runner error, got %v", err)
	}
	if len(pager.calledWith) != 0 {
		t.Fatalf("pager should not be called")
	}
}

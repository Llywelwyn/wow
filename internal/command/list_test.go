package command

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/model"
	"github.com/llywelwyn/wow/internal/storage"
)

func newListCommand(t *testing.T, metas []model.Metadata) (*ListCommand, func()) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "meta.db")
	db, err := storage.InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}

	ctx := context.Background()
	for _, m := range metas {
		if err := storage.InsertMetadata(ctx, db, m); err != nil {
			t.Fatalf("InsertMetadata(%q) error = %v", m.Key, err)
		}
	}

	cmd := &ListCommand{
		DB:     db,
		Output: io.Discard,
	}

	cleanup := func() {
		_ = db.Close()
	}

	return cmd, cleanup
}

func TestListCommandDefaultPlain(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cmd, cleanup := newListCommand(t, []model.Metadata{
		{Key: "first", Created: now.Add(-time.Hour), Modified: now.Add(-time.Hour)},
		{Key: "second", Created: now, Modified: now},
	})
	defer cleanup()

	var out bytes.Buffer
	cmd.Output = &out

	if err := cmd.Execute(nil); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	output := strings.TrimSuffix(out.String(), "\n")
	lines := strings.Split(output, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d (%q)", len(lines), lines)
	}
	if lines[0] != "second" {
		t.Fatalf("expected newest entry first, got %q", lines[0])
	}
	if lines[1] != "first" {
		t.Fatalf("expected oldest entry second, got %q", lines[1])
	}
}

func TestListCommandLimitAndPage(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	cmd, cleanup := newListCommand(t, []model.Metadata{
		{Key: "first", Created: now.Add(-2 * time.Hour), Modified: now.Add(-2 * time.Hour)},
		{Key: "second", Created: now.Add(-time.Hour), Modified: now.Add(-time.Hour)},
		{Key: "third", Created: now, Modified: now},
	})
	defer cleanup()

	var out bytes.Buffer
	cmd.Output = &out

	if err := cmd.Execute([]string{"--limit", "2"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d (%q)", len(lines), lines)
	}
	if lines[0] != "third" || lines[1] != "second" {
		t.Fatalf("unexpected order: %q", lines)
	}

	out.Reset()
	if err := cmd.Execute([]string{"--limit", "1", "--page", "2"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	lines = strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 1 || lines[0] != "second" {
		t.Fatalf("expected page 2 to show second, got %q", lines)
	}
}

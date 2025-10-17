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

func TestListCommandPlainDefault(t *testing.T) {
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

func TestListCommandVerbose(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	cmd, cleanup := newListCommand(t, []model.Metadata{
		{
			Key:         "foo",
			Type:        "text",
			Created:     now,
			Modified:    now,
			Description: "desc",
			Tags:        "a,b",
		},
	})
	defer cleanup()

	var out bytes.Buffer
	cmd.Output = &out

	if err := cmd.Execute([]string{"--plain", "--verbose"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	got := strings.TrimSuffix(out.String(), "\n")
	fields := strings.Split(got, "\t")
	if len(fields) != 5 {
		t.Fatalf("expected 5 fields, got %d (%q)", len(fields), fields)
	}
	if fields[0] != "foo" {
		t.Fatalf("unexpected key field: %q", fields[0])
	}
	if fields[3] != "a,b" {
		t.Fatalf("unexpected tags: %q", fields[3])
	}
	if fields[4] != "desc" {
		t.Fatalf("unexpected description: %q", fields[4])
	}
	if fields[1] != now.Format(time.RFC3339) {
		t.Fatalf("expected created timestamp %q, got %q", now.Format(time.RFC3339), fields[1])
	}
}

func TestListCommandCustomDelimiter(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cmd, cleanup := newListCommand(t, []model.Metadata{
		{Key: "first", Created: now, Modified: now, Tags: "one"},
	})
	defer cleanup()

	var out bytes.Buffer
	cmd.Output = &out

	if err := cmd.Execute([]string{"--plain=|"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	line := strings.TrimSuffix(out.String(), "\n")
	if line != "first" {
		t.Fatalf("unexpected output: %q", line)
	}
}

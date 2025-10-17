package command

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/services"
	"github.com/llywelwyn/wow/internal/storage"
)

func TestGetCommandReadsSnippet(t *testing.T) {
	base := t.TempDir()
	dbPath := filepath.Join(base, "meta.db")
	db, err := storage.InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	defer db.Close()

	saver := &services.Saver{
		BaseDir: base,
		DB:      db,
		Now: func() time.Time {
			return time.Unix(1_700_000_000, 0)
		},
	}

	cmd := &SaveCommand{
		Saver:  saver,
		Input:  strings.NewReader("hello world"),
		Output: io.Discard,
	}
	if err := cmd.Execute([]string{"go/foo"}); err != nil {
		t.Fatalf("Save Execute error = %v", err)
	}

	var out bytes.Buffer
	getCmd := &GetCommand{
		BaseDir: base,
		Output:  &out,
	}

	if err := getCmd.Execute([]string{"go/foo"}); err != nil {
		t.Fatalf("Get Execute error = %v", err)
	}

	if out.String() != "hello world" {
		t.Fatalf("output = %q, want %q", out.String(), "hello world")
	}
}

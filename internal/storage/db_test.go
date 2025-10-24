package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitMetaDBCreatesSchema(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "nested", ".meta.db")

	db, err := InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	var table string
	if err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='snippets'").Scan(&table); err != nil {
		t.Fatalf("schema query error = %v", err)
	}
	if table != "snippets" {
		t.Fatalf("table name = %q, want snippets", table)
	}

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected db file to exist: %v", err)
	}
}

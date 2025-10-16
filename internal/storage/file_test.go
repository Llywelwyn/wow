package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveReadDeleteLifecycle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snippets", "foo")

	want := []byte("hello world")
	if err := Save(path, bytes.NewReader(want)); err != nil {
		t.Fatalf("Save error = %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("Read = %q, want %q", got, want)
	}

	if err := Delete(path); err != nil {
		t.Fatalf("Delete error = %v", err)
	}

	if _, err := Read(path); err != ErrNotFound {
		t.Fatalf("Read after delete err = %v, want ErrNotFound", err)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "foo")

	exists, err := Exists(path)
	if err != nil {
		t.Fatalf("Exists error = %v", err)
	}
	if exists {
		t.Fatalf("Exists should be false for missing file")
	}

	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	exists, err = Exists(path)
	if err != nil {
		t.Fatalf("Exists error = %v", err)
	}
	if !exists {
		t.Fatalf("Exists should be true after writing")
	}
}

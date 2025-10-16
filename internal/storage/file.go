package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ErrNotFound is returned when the snippet file is missing.
var ErrNotFound = os.ErrNotExist

// Save writes content to the given path using an atomic workflow.
// It ensures parent directories exist and applies 0600 permissions.
func Save(path string, content io.Reader) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create snippet dir %q: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".wow-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if err := tmp.Chmod(0o600); err != nil {
		cleanup()
		return fmt.Errorf("chmod temp file: %w", err)
	}

	if _, err := io.Copy(tmp, content); err != nil {
		cleanup()
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

// Read returns the contents of the snippet file at the given path.
func Read(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read snippet file: %w", err)
	}
	return data, nil
}

// Delete removes the snippet file at the given path.
func Delete(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("delete snippet file: %w", err)
	}
	return nil
}

// Exists reports whether the path already exists on disk.
func Exists(path string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat snippet file: %w", err)
	}
	return !info.IsDir(), nil
}

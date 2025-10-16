// config manages configuration.
// It captures required paths from ENV.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config stores wow base directory
// and the metadata DB paths.
type Config struct {
	BaseDir string
	MetaDB  string
}

// Load resolves configuration from environment
// and ensures required directories exist.
func Load() (Config, error) {
	base, err := resolveBaseDir()
	if err != nil {
		return Config{}, err
	}

	if err := os.MkdirAll(base, 0o700); err != nil {
		return Config{}, fmt.Errorf("create base dir %q: %w", base, err)
	}

	return Config{
		BaseDir: base,
		MetaDB:  filepath.Join(base, "meta.db"),
	}, nil
}

// resolveBaseDir figures out the base directory we want to use.
//
// It falls back through the following directories in order:
//   - $WOW_HOME
//   - $XDG_DATA_HOME
//   - $HOME
//
// It returns the first directory to resolve.
// If no fallbacks resolve a valid directory, it errors.
func resolveBaseDir() (string, error) {
	if dir := strings.TrimSpace(os.Getenv("WOW_HOME")); dir != "" {
		return normalizeDir(dir)
	}

	if xdg := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); xdg != "" {
		return normalizeDir(filepath.Join(xdg, "wow"))
	}

	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", errors.New("cannot determine home directory; set WOW_HOME explicitly")
	}

	return normalizeDir(filepath.Join(home, ".wow"))
}

// normalizeDir normalises a directory string to an absolute filepath.
// If a path fails to be resolved, it errors.
func normalizeDir(dir string) (string, error) {
	expanded, err := expandHome(dir)
	if err != nil {
		return "", err
	}

	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", expanded, err)
	}

	return abs, nil
}

// expandHome returns a given path with preceding tilde replaced with $HOME.
// If $HOME does not exist, it errors.
func expandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("expand home for %q: %w", path, err)
	}

	if path == "~" {
		return home, nil
	}

	trimmed := strings.TrimPrefix(path, "~")
	return filepath.Join(home, strings.TrimPrefix(trimmed, string(filepath.Separator))), nil
}

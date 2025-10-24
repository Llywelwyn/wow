// config manages configuration.
// It captures required paths from ENV.
package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/llywelwyn/wow/internal/editor"
	"github.com/llywelwyn/wow/internal/opener"
	"github.com/llywelwyn/wow/internal/runner"
	"github.com/llywelwyn/wow/internal/storage"
)

// Config stores wow base directory
// and the metadata DB paths.
type Config struct {
	BaseDir string
	DB      *sql.DB
	idFile  string
	Input   io.Reader
	Output  io.Writer
	Clock   func() time.Time
	Editor  func(context.Context, string) error
	Opener  func(context.Context, string) error
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

	db, err := storage.InitMetaDB(filepath.Join(base, ".meta.db"))
	if err != nil {
		return Config{}, fmt.Errorf("init meta db: %w", err)
	}

	cfg := Config{
		BaseDir: base,
		DB:      db,
		idFile:  filepath.Join(base, ".id"),
		Input:   os.Stdin,
		Output:  os.Stdout,
		Clock:   time.Now,
		Editor:  runner.Run(editor.GetEditorFromEnv()),
		Opener:  runner.Run(opener.GetOpenerFromEnv()),
	}

	if err := cfg.initIdFile(); err != nil {
		return Config{}, err
	}
	return cfg, nil
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
// Currently does not handle expanding home of other users. e.g. ~username
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

func (c *Config) initIdFile() error {
	if _, err := os.Stat(c.idFile); os.IsNotExist(err) {
		return os.WriteFile(c.idFile, []byte("1\n"), 0o600)
	}
	return nil
}

func (c *Config) NextId() (int, error) {
	data, err := os.ReadFile(c.idFile)
	if err != nil {
		return 0, fmt.Errorf("read ID file: %w", err)
	}

	idstr := strings.TrimSpace(string(data))
	currentId, err := strconv.Atoi(idstr)
	if err != nil {
		return 0, fmt.Errorf("parse current ID %q: %w", idstr, err)
	}

	tryId := currentId
	for {
		path := filepath.Join(c.BaseDir, strconv.Itoa(tryId))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
		tryId++
	}

	nextId := tryId + 1
	if err := os.WriteFile(c.idFile, fmt.Appendf(nil, "%d\n", nextId), 0o600); err != nil {
		return 0, fmt.Errorf("write next ID %d: %w", nextId, err)
	}

	return currentId, nil
}

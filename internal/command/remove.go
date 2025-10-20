package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/storage"
)

// RemoveCommand deletes snippets identified by key.
type RemoveCommand struct {
	BaseDir string
	DB      *sql.DB
}

// NewRemoveCommand constructs a RemoveCommand using defaults from cfg.
func NewRemoveCommand(cfg Config) *RemoveCommand {
	return &RemoveCommand{
		BaseDir: cfg.BaseDir,
		DB:      cfg.DB,
	}
}

// Name returns the command keyword.
func (c *RemoveCommand) Name() string { return "remove" }

// Execute deletes the provided snippet key.
func (c *RemoveCommand) Execute(args []string) error {
	if c.DB == nil || strings.TrimSpace(c.BaseDir) == "" {
		return errors.New("remove command not configured")
	}

	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	var help *bool = fs.BoolP("help", "h", false, "display help")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *help {
		fmt.Fprintln(os.Stdout, `Usage:
  wow remove <key>`)
		fs.PrintDefaults()
		return nil
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		return errors.New("key required")
	}
	key := remaining[0]
	return c.removeSnippet(context.Background(), key)
}

func (c *RemoveCommand) removeSnippet(ctx context.Context, rawKey string) error {
	if c.DB == nil {
		return errors.New("remove command not configured")
	}

	normalized, err := key.Normalize(rawKey)
	if err != nil {
		return err
	}

	if err := storage.DeleteMetadata(ctx, c.DB, normalized); err != nil {
		return err
	}

	path, err := key.ResolvePath(c.BaseDir, normalized)
	if err != nil {
		return err
	}

	if err := storage.Delete(path); err != nil && !errors.Is(err, storage.ErrNotFound) {
		return err
	}

	return nil
}

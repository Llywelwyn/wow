package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/model"
	"github.com/llywelwyn/wow/internal/storage"
)

// EditCommand opens an existing snippet in the user's editor.
type EditCommand struct {
	BaseDir string
	DB      *sql.DB
	Now     func() time.Time
	Open    func(context.Context, string) error
}

// NewEditCommand constructs an EditCommand using defaults from cfg.
func NewEditCommand(cfg Config) *EditCommand {
	return &EditCommand{
		BaseDir: cfg.BaseDir,
		DB:      cfg.DB,
		Now:     cfg.clock(),
		Open:    cfg.editor(),
	}
}

// Name returns the command keyword.
func (c *EditCommand) Name() string { return "edit" }

// Execute edits the snippet identified by key.
func (c *EditCommand) Execute(args []string) error {
	if c.DB == nil || c.Now == nil || c.Open == nil || strings.TrimSpace(c.BaseDir) == "" {
		return errors.New("edit command not configured")
	}

	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	var help *bool = fs.BoolP("help", "h", false, "display help")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *help {
		fmt.Fprintln(os.Stdout, `Usage:
  wow edit <key>`)
		fs.PrintDefaults()
		return nil
	}

	remaining := fs.Args()
	if len(remaining) != 1 {
		return errors.New("edit expects exactly one key")
	}
	_, err := c.editSnippet(context.Background(), remaining[0])
	return err
}

func (c *EditCommand) editSnippet(ctx context.Context, rawKey string) (model.Metadata, error) {
	if c.DB == nil || c.Now == nil || c.Open == nil {
		return model.Metadata{}, errors.New("edit command not configured")
	}

	normalized, err := key.Normalize(rawKey)
	if err != nil {
		return model.Metadata{}, err
	}

	meta, err := storage.GetMetadata(ctx, c.DB, normalized)
	if err != nil {
		return model.Metadata{}, err
	}

	path, err := key.ResolvePath(c.BaseDir, normalized)
	if err != nil {
		return model.Metadata{}, err
	}

	before, err := os.Stat(path)
	if err != nil {
		return model.Metadata{}, err
	}

	if err := c.Open(ctx, path); err != nil {
		return model.Metadata{}, err
	}

	after, err := os.Stat(path)
	if err != nil {
		return model.Metadata{}, err
	}

	changed := after.ModTime() != before.ModTime() || after.Size() != before.Size()
	if !changed {
		return meta, nil
	}

	data, err := storage.Read(path)
	if err != nil {
		return model.Metadata{}, err
	}

	meta.Type = detectSnippetType(data)
	meta.Modified = c.Now().UTC()

	if err := storage.UpdateMetadata(ctx, c.DB, meta); err != nil {
		return model.Metadata{}, err
	}

	return meta, nil
}

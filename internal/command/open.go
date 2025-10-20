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

// OpenOptions controls how a snippet should be opened.
type OpenOptions struct {
	UsePager bool
}

// OpenCommand launches snippets via configured opener or pager.
type OpenCommand struct {
	BaseDir   string
	DB        *sql.DB
	OpenFunc  func(context.Context, string) error
	PagerFunc func(context.Context, string) error
}

// NewOpenCommand constructs an OpenCommand using defaults from cfg.
func NewOpenCommand(cfg Config) *OpenCommand {
	return &OpenCommand{
		BaseDir:   cfg.BaseDir,
		DB:        cfg.DB,
		OpenFunc:  cfg.opener(),
		PagerFunc: cfg.pager(),
	}
}

// Name returns the command keyword.
func (c *OpenCommand) Name() string { return "open" }

// Execute opens the snippet identified by key.
func (c *OpenCommand) Execute(args []string) error {
	if c.DB == nil || c.OpenFunc == nil || c.PagerFunc == nil || strings.TrimSpace(c.BaseDir) == "" {
		return errors.New("open command not configured")
	}

	fs := flag.NewFlagSet("open", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	var pager *bool = fs.BoolP("pager", "p", false, "view snippet in pager")
	var help *bool = fs.BoolP("help", "h", false, "display help")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *help {
		fmt.Fprintln(os.Stdout, `Usage:
  wow open <key> [--pager]`)
		fs.PrintDefaults()
		return nil
	}

	remaining := fs.Args()
	if len(remaining) != 1 {
		return errors.New("open expects exactly one key")
	}

	return c.openSnippet(context.Background(), remaining[0], OpenOptions{UsePager: *pager})
}

func (c *OpenCommand) openSnippet(ctx context.Context, rawKey string, opts OpenOptions) error {
	if c.DB == nil || c.OpenFunc == nil || c.PagerFunc == nil {
		return errors.New("open command not configured")
	}

	normalized, err := key.Normalize(rawKey)
	if err != nil {
		return err
	}

	meta, err := storage.GetMetadata(ctx, c.DB, normalized)
	if err != nil {
		return err
	}

	path, err := key.ResolvePath(c.BaseDir, normalized)
	if err != nil {
		return err
	}

	if opts.UsePager {
		return c.PagerFunc(ctx, path)
	}

	if meta.Type == "url" {
		data, err := storage.Read(path)
		if err != nil {
			return err
		}
		target := firstNonEmptyLine(data)
		if target == "" {
			target = string(data)
		}
		target = strings.TrimSpace(target)
		if target == "" {
			target = path
		}
		return c.OpenFunc(ctx, target)
	}

	return c.OpenFunc(ctx, path)
}

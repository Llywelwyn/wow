package command

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/llywelwyn/wow/internal/services"
	flag "github.com/spf13/pflag"
)

type openHandler interface {
	Open(ctx context.Context, key string, opts services.OpenOptions) error
}

// OpenCommand launches snippets via configured opener or pager.
type OpenCommand struct {
	Opener openHandler
}

// NewOpenCommand constructs an OpenCommand using defaults from cfg.
func NewOpenCommand(cfg Config) *OpenCommand {
	return &OpenCommand{
		Opener: &services.Opener{
			BaseDir:   cfg.BaseDir,
			DB:        cfg.DB,
			OpenFunc:  cfg.opener(),
			PagerFunc: cfg.pager(),
		},
	}
}

// Name returns the command keyword.
func (c *OpenCommand) Name() string { return "open" }

// Execute opens the snippet identified by key.
func (c *OpenCommand) Execute(args []string) error {
	if c.Opener == nil {
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

	return c.Opener.Open(context.Background(), remaining[0], services.OpenOptions{UsePager: *pager})
}

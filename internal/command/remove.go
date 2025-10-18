package command

import (
	"context"
	"errors"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"

	"github.com/llywelwyn/wow/internal/services"
)

// RemoveCommand deletes snippets identified by key.
type RemoveCommand struct {
	Remover *services.Remover
}

// NewRemoveCommand constructs a RemoveCommand using defaults from cfg.
func NewRemoveCommand(cfg Config) *RemoveCommand {
	return &RemoveCommand{
		Remover: &services.Remover{
			BaseDir: cfg.BaseDir,
			DB:      cfg.DB,
		},
	}
}

// Name returns the command keyword.
func (c *RemoveCommand) Name() string { return "remove" }

// Execute deletes the provided snippet key.
func (c *RemoveCommand) Execute(args []string) error {
	if c.Remover == nil {
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
	return c.Remover.Remove(context.Background(), key)
}

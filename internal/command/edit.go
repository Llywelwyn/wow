package command

import (
	"context"
	"errors"

	"github.com/llywelwyn/wow/internal/model"
)

// EditHandler wraps the edit behaviour required by the command.
type EditHandler interface {
	Edit(ctx context.Context, key string) (model.Metadata, error)
}

// EditCommand opens an existing snippet in the user's editor.
type EditCommand struct {
	Editor EditHandler
}

// Name returns the command keyword.
func (c *EditCommand) Name() string { return "edit" }

// Execute edits the snippet identified by key.
func (c *EditCommand) Execute(args []string) error {
	if c.Editor == nil {
		return errors.New("edit command not configured")
	}
	if len(args) != 1 {
		return errors.New("edit expects exactly one key")
	}
	_, err := c.Editor.Edit(context.Background(), args[0])
	return err
}

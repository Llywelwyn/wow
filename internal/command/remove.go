package command

import (
	"context"
	"errors"

	"github.com/llywelwyn/wow/internal/core"
)

// RemoveCommand deletes snippets identified by key.
type RemoveCommand struct {
	Remover *core.Remover
}

// Name returns the command keyword.
func (c *RemoveCommand) Name() string { return "remove" }

// Execute deletes the provided snippet key.
func (c *RemoveCommand) Execute(args []string) error {
	if c.Remover == nil {
		return errors.New("remove command not configured")
	}
	if len(args) == 0 {
		return errors.New("key required")
	}
	key := args[0]
	return c.Remover.Remove(context.Background(), key)
}

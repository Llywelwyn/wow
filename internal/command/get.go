package command

import (
	"errors"
	"fmt"
	"io"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/storage"
)

// GetCommand streams snippet content to stdout.
type GetCommand struct {
	BaseDir string
	Output  io.Writer
}

// Name returns the keyword for explicit invocation.
func (c *GetCommand) Name() string {
	return "get"
}

// Execute reads the snippet file resolved by the provided key and writes it to output.
func (c *GetCommand) Execute(args []string) error {
	if c.Output == nil || c.BaseDir == "" {
		return errors.New("get command not fully configured")
	}

	if len(args) == 0 {
		return errors.New("key required")
	}

	rawKey := args[0]
	normalized, err := key.Normalize(rawKey)
	if err != nil {
		return err
	}

	path, err := key.ResolvePath(c.BaseDir, normalized)
	if err != nil {
		return err
	}

	data, err := storage.Read(path)
	if err != nil {
		return err
	}

	if _, err := c.Output.Write(data); err != nil {
		return fmt.Errorf("write snippet to output: %w", err)
	}
	return nil
}

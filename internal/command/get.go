package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/services"
	"github.com/llywelwyn/wow/internal/storage"
)

// GetCommand streams snippet content to stdout and optionally mutates tags.
type GetCommand struct {
	BaseDir string
	Output  io.Writer
	Meta    *services.Metadata
}

// NewGetCommand constructs a GetCommand using defaults from cfg.
func NewGetCommand(cfg Config) *GetCommand {
	var meta *services.Metadata
	if cfg.DB != nil {
		meta = &services.Metadata{
			DB:  cfg.DB,
			Now: cfg.clock(),
		}
	}
	return &GetCommand{
		BaseDir: cfg.BaseDir,
		Output:  cfg.writer(),
		Meta:    meta,
	}
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

	tagArgs := extractTagArgs(args)
	if len(tagArgs.Others) == 0 {
		return errors.New("key required")
	}

	keyArg := tagArgs.Others[0]
	flagArgs := tagArgs.Others[1:]

	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	var addCSV, removeCSV string
	fs.StringVar(&addCSV, "tag", "", "comma-separated tags to add")
	fs.StringVar(&removeCSV, "untag", "", "comma-separated tags to remove")
	fs.SetOutput(io.Discard)

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	path, err := key.ResolvePath(c.BaseDir, keyArg)
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

	addTags := append(splitTags(addCSV), tagArgs.Add...)
	removeTags := append(splitTags(removeCSV), tagArgs.Remove...)
	if c.Meta != nil && (len(addTags) > 0 || len(removeTags) > 0) {
		if _, err := c.Meta.UpdateTags(context.Background(), keyArg, addTags, removeTags); err != nil {
			return err
		}
	}

	return nil
}

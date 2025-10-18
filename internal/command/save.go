package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/llywelwyn/wow/internal/services"
)

// SaveCommand persists snippet content read from stdin and prints the resolved key.
type SaveCommand struct {
	Saver  *services.Saver
	Input  io.Reader
	Output io.Writer
}

// NewSaveCommand constructs a SaveCommand using default dependencies from cfg.
func NewSaveCommand(cfg Config) *SaveCommand {
	return &SaveCommand{
		Saver: &services.Saver{
			BaseDir: cfg.BaseDir,
			DB:      cfg.DB,
			Now:     cfg.clock(),
		},
		Input:  cfg.reader(),
		Output: cfg.writer(),
	}
}

// Name returns the command keyword for explicit invocation.
func (c *SaveCommand) Name() string {
	return "save"
}

// Execute saves the snippet using the provided arguments.
func (c *SaveCommand) Execute(args []string) error {
	if c.Saver == nil || c.Input == nil || c.Output == nil {
		return errors.New("save command not fully configured")
	}

	tagArgs := extractTagArgs(args)
	args = tagArgs.Others

	fs := flag.NewFlagSet("save", flag.ContinueOnError)
	desc := fs.String("desc", "", "description")
	var tagsCSV string
	fs.StringVar(&tagsCSV, "tag", "", "comma-separated tags")
	fs.SetOutput(io.Discard)

	var keyArg string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		keyArg = args[0]
		args = args[1:]
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	addTags := append(splitTags(tagsCSV), tagArgs.Add...)

	res, err := c.Saver.Save(context.Background(), services.SaveRequest{
		Key:         keyArg,
		Description: *desc,
		Tags:        addTags,
		Reader:      c.Input,
	})
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(c.Output, res.Key); err != nil {
		return fmt.Errorf("write key to output: %w", err)
	}
	return nil
}

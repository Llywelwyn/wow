package command

import (
	"context"
	"errors"
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"strings"

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

	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	fs.SetOutput(c.Output)
	var addCSV *string = fs.StringP("tag", "t", "", "comma-separated tags to add")
	var removeCSV *string = fs.StringP("untag", "u", "", "comma-separated tags to remove")
	var help *bool = fs.BoolP("help", "h", false, "display help")
	usage := `Usage:
  wow get <key> [--tag tags] [--untag tags] [@tag ...] [-@tag ...]`

	if len(tagArgs.Others) == 0 {
		return errors.New("key required")
	}

	keyArg := tagArgs.Others[0]
	if strings.HasPrefix(keyArg, "-") {
		if err := fs.Parse(tagArgs.Others); err != nil {
			return err
		}
		if *help {
			fmt.Fprintln(c.Output, usage)
			fs.PrintDefaults()
			return nil
		}
		return errors.New("key must be the first argument")
	}

	if err := fs.Parse(tagArgs.Others[1:]); err != nil {
		return err
	}

	if *help {
		fmt.Fprintln(c.Output, usage)
		fs.PrintDefaults()
		return nil
	}

	addTags := append(splitTags(*addCSV), tagArgs.Add...)
	removeTags := append(splitTags(*removeCSV), tagArgs.Remove...)
	hasTagChange := len(addTags) > 0 || len(removeTags) > 0

	path, err := key.ResolvePath(c.BaseDir, keyArg)
	if err != nil {
		return err
	}

	if !hasTagChange {
		data, err := storage.Read(path)
		if err != nil {
			return err
		}
		if _, err := c.Output.Write(data); err != nil {
			return fmt.Errorf("write snippet to output: %w", err)
		}
		return nil
	}

	if c.Meta == nil {
		return errors.New("metadata updates not supported")
	}

	result, err := c.Meta.UpdateTags(context.Background(), keyArg, addTags, removeTags)
	if err != nil {
		return err
	}

	return writeTagSummary(c.Output, result.Added, result.Removed)
}

func writeTagSummary(w io.Writer, added, removed []string) error {
	if len(added) == 0 && len(removed) == 0 {
		_, err := fmt.Fprintln(w, "tags unchanged")
		return err
	}
	if len(added) > 0 {
		if _, err := fmt.Fprintf(w, "added %s\n", formatTagList(added)); err != nil {
			return err
		}
	}
	if len(removed) > 0 {
		if _, err := fmt.Fprintf(w, "removed %s\n", formatTagList(removed)); err != nil {
			return err
		}
	}
	return nil
}

func formatTagList(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	formatted := make([]string, len(tags))
	for i, tag := range tags {
		formatted[i] = "@" + tag
	}
	return strings.Join(formatted, " ")
}

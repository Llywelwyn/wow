package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"golang.org/x/term"

	"github.com/llywelwyn/wow/internal/model"
	"github.com/llywelwyn/wow/internal/storage"
)

// ListCommand prints snippet metadata.
type ListCommand struct {
	DB     *sql.DB
	Output io.Writer
}

// NewListCommand constructs a ListCommand using defaults from cfg.
func NewListCommand(cfg Config) *ListCommand {
	return &ListCommand{
		DB:     cfg.DB,
		Output: cfg.writer(),
	}
}

// Name returns the command keyword for invocation.
func (c *ListCommand) Name() string {
	return "list"
}

// Execute lists metadata rows according to the provided flags.
func (c *ListCommand) Execute(args []string) error {
	if c.DB == nil || c.Output == nil {
		return errors.New("list command not fully configured")
	}

	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.SetOutput(c.Output)
	var plain *string = fs.StringP("plain", "p", "", "removes pretty formatting; pass a string to override tab-delimiter")
	fs.Lookup("plain").NoOptDefVal = "\t"
	var quiet *bool = fs.BoolP("quiet", "q", false, "quiet output: only keys")
	var help *bool = fs.BoolP("help", "h", false, "display help")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *help {
		fmt.Fprintln(c.Output, `Usage:
  wow list [--plain[=delimiter]] [--quiet]

  Wow! Lists metadata for all the snippets you've got saved.

  By default, output is pretty-printed as a table. When you
  pipe the stdout elsewhere or pass --plain, the output is
  tab-delimited and there's no header row to make scripting
  easier. If you want to use a different delimiter, you can
  pass it as an argument to --plain.`)
		fmt.Fprintln(c.Output)
		fs.PrintDefaults()
		return nil
	}

	ctx := context.Background()
	entries, err := storage.ListMetadata(ctx, c.DB)
	if err != nil {
		return err
	}

	if *plain != "" || !writerIsTerminal(c.Output) {
		delimiter := *plain
		if delimiter == "" {
			delimiter = "\t"
		}
		return renderPlain(c.Output, entries, *quiet, delimiter)
	}
	return renderPretty(c.Output, entries, *quiet)
}

func writerIsTerminal(w io.Writer) bool {
	type fdWriter interface {
		io.Writer
		Fd() uintptr
	}
	if f, ok := w.(fdWriter); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

func renderPlain(w io.Writer, entries []model.Metadata, quiet bool, delimiter string) error {
	for _, meta := range entries {
		var fields []string
		if quiet {
			fields = []string{meta.Key}
		} else {
			fields = []string{
				meta.Key,
				meta.Tags,
				meta.Description,
				meta.Created.UTC().Format(time.RFC3339),
				meta.Modified.UTC().Format(time.RFC3339),
			}
		}

		if _, err := fmt.Fprintln(w, strings.Join(fields, delimiter)); err != nil {
			return err
		}
	}
	return nil
}

func renderPretty(w io.Writer, entries []model.Metadata, quiet bool) error {
	tw := tabwriter.NewWriter(w, 2, 4, 2, ' ', 0)
	if quiet {
		fmt.Fprintln(tw, "KEY")
	} else {
		fmt.Fprintln(tw, "KEY\tTAGS\tMODIFIED\tCREATED\tDESCRIPTION")
	}

	for _, meta := range entries {
		if quiet {
			fmt.Fprintf(tw, "%s\n", meta.Key)
		} else {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
				meta.Key,
				meta.Tags,
				meta.Created.UTC().Format(time.RFC3339),
				meta.Modified.UTC().Format(time.RFC3339),
				meta.Description,
			)
		}
	}

	if err := tw.Flush(); err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Fprintln(w, "(no snippets)")
	}
	return nil
}

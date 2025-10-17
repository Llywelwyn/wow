package command

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
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
	fs.SetOutput(io.Discard)
	plain := newPlainFlag()
	fs.Var(plain, "plain", "plain output; optionally provide a delimiter (default: tab)")
	verbose := fs.Bool("v", false, "verbose output")
	fs.BoolVar(verbose, "verbose", false, "verbose output")

	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx := context.Background()
	entries, err := storage.ListMetadata(ctx, c.DB)
	if err != nil {
		return err
	}

	modePlain := plain.Requested() || !writerIsTerminal(c.Output)
	delimiter := plain.Delimiter()
	if modePlain {
		return renderPlain(c.Output, entries, *verbose, delimiter)
	}
	return renderPretty(c.Output, entries, *verbose)
}

type plainFlag struct {
	requested bool
	delimiter string
}

func newPlainFlag() *plainFlag {
	return &plainFlag{
		delimiter: "\t",
	}
}

func (p *plainFlag) String() string {
	if !p.requested {
		return ""
	}
	return p.delimiter
}

func (p *plainFlag) Set(value string) error {
	switch value {
	case "", "true":
		p.requested = true
		p.delimiter = "\t"
		return nil
	case "false":
		p.requested = false
		p.delimiter = "\t"
		return nil
	}

	del, err := parseDelimiter(value)
	if err != nil {
		return err
	}

	p.requested = true
	p.delimiter = del
	return nil
}

func (p *plainFlag) IsBoolFlag() bool {
	return true
}

func (p *plainFlag) Requested() bool {
	return p.requested
}

func (p *plainFlag) Delimiter() string {
	return p.delimiter
}

func parseDelimiter(val string) (string, error) {
	if val == "" {
		return "\t", nil
	}

	if parsed, err := strconv.Unquote(`"` + val + `"`); err == nil {
		return parsed, nil
	}

	return val, nil
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

func renderPlain(w io.Writer, entries []model.Metadata, verbose bool, delimiter string) error {
	for _, meta := range entries {
		var fields []string
		if verbose {
			fields = []string{
				meta.Key,
				meta.Created.UTC().Format(time.RFC3339),
				meta.Modified.UTC().Format(time.RFC3339),
				meta.Tags,
				meta.Description,
			}
		} else {
			fields = []string{meta.Key}
		}
		if _, err := fmt.Fprintln(w, strings.Join(fields, delimiter)); err != nil {
			return err
		}
	}
	return nil
}

func renderPretty(w io.Writer, entries []model.Metadata, verbose bool) error {
	tw := tabwriter.NewWriter(w, 2, 4, 2, ' ', 0)
	if verbose {
		fmt.Fprintln(tw, "KEY\tCREATED\tMODIFIED\tTAGS\tDESCRIPTION")
	} else {
		fmt.Fprintln(tw, "KEY\tTAGS")
	}

	for _, meta := range entries {
		if verbose {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
				meta.Key,
				meta.Created.UTC().Format(time.RFC3339),
				meta.Modified.UTC().Format(time.RFC3339),
				meta.Tags,
				meta.Description,
			)
		} else {
			fmt.Fprintf(tw, "%s\t%s\n", meta.Key, meta.Tags)
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

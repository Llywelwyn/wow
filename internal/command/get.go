package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/services"
	"github.com/llywelwyn/wow/internal/storage"
	"github.com/llywelwyn/wow/internal/ui"
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

	if len(tagArgs.Others) == 0 {
		return errors.New("key required")
	}

	keyArg := tagArgs.Others[0]
	if strings.HasPrefix(keyArg, "-") {
		if err := fs.Parse(tagArgs.Others); err != nil {
			return err
		}
		// TODO: This is completely duplicated. Dedupe this with a refactor of parsing --help on implicit Gets.
		if *help {
			fmt.Fprintln(c.Output, `Usage:
  wow get <key> [--tag tag1,tag2] [--untag tag1] [@tag1 @tag2] [-@tag1]

  wow! Fetches a snippet, or modifies its metadata.

  You can get implicitly by running "wow <key>"
  without the "get" keyword. Provided no input
  was piped in, wow! guesses you want to fetch.

  If you add any flags, rather than outputting
  the content of the snippet, wow! will update
  the metadata instead. 

  Some examples:
    wow foo             -->  fetches the content of "foo".
    wow foo @bar -@baz  -->  adds "bar" and removes "baz" from tags.
    wow foo --tag 1,2   -->  adds "1" and "2" to tags.`)
			fmt.Fprintln(c.Output)
			fs.PrintDefaults()
			fmt.Fprintln(c.Output, `
  PS.
    You wont be able to implicitly get if your
    key collides with another command. In that
    case, you need to specify: "wow get <key>"`)
			return nil
		}
		return errors.New("key must be the first argument")
	}

	if err := fs.Parse(tagArgs.Others[1:]); err != nil {
		return err
	}

	if *help {
		fmt.Fprintln(c.Output, `Usage:
  wow get <key> [--tag tag1,tag2] [--untag tag1] [@tag1 @tag2] [-@tag1]

  wow! Fetches a snippet, or modifies its metadata.

  You can get implicitly by running "wow <key>"
  without the "get" keyword. Provided no input
  was piped in, wow! guesses you want to fetch.

  If you add any flags, rather than outputting
  the content of the snippet, wow! will update
  the metadata instead. 

  Some examples:
    wow foo             -->  fetches the content of "foo".
    wow foo @bar -@baz  -->  adds "bar" and removes "baz" from tags.
	wow foo --tag 1,2   -->  adds "1" and "2" to tags.`)
		fmt.Fprintln(c.Output)
		fs.PrintDefaults()
		fmt.Fprintln(c.Output, `
  PS.
    You wont be able to implicitly get if your
    key collides with another command. In that
    case, you need to specify: "wow get <key>"`)
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
	styles := ui.DefaultStyles()

	if len(added) == 0 && len(removed) == 0 {
		_, err := fmt.Fprintln(w, styles.Subtle.Render("tags unchanged"))
		return err
	}
	if len(added) > 0 {
		if _, err := fmt.Fprintf(w, "%s %s\n", styles.Positive.Render("added"), formatTagList(added)); err != nil {
			return err
		}
	}
	if len(removed) > 0 {
		if _, err := fmt.Fprintf(w, "%s %s\n", styles.Negative.Render("removed"), formatTagList(removed)); err != nil {
			return err
		}
	}
	return nil
}

func formatTagList(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	styles := ui.DefaultStyles()

	formatted := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		formatted = append(formatted, styles.Tag.Render("@"+tag))
	}
	return strings.Join(formatted, " ")
}

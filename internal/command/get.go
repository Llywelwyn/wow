package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/storage"
	"github.com/llywelwyn/wow/internal/ui"
)

// GetCommand streams snippet content to stdout and optionally mutates tags.
type GetCommand struct {
	BaseDir string
	DB      *sql.DB
	Output  io.Writer
	Now     func() time.Time
}

// NewGetCommand constructs a GetCommand using defaults from cfg.
func NewGetCommand(cfg Config) *GetCommand {
	return &GetCommand{
		BaseDir: cfg.BaseDir,
		DB:      cfg.DB,
		Output:  cfg.writer(),
		Now:     cfg.clock(),
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

	if c.DB == nil || c.Now == nil {
		return errors.New("metadata updates not supported")
	}

	added, removed, err := c.updateTags(context.Background(), keyArg, addTags, removeTags)
	if err != nil {
		return err
	}

	return writeTagSummary(c.Output, added, removed)
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

func (c *GetCommand) updateTags(ctx context.Context, rawKey string, add, remove []string) ([]string, []string, error) {
	if c.DB == nil || c.Now == nil {
		return nil, nil, errors.New("metadata updates not supported")
	}

	normalized, err := key.Normalize(rawKey)
	if err != nil {
		return nil, nil, err
	}

	meta, err := storage.GetMetadata(ctx, c.DB, normalized)
	if err != nil {
		return nil, nil, err
	}

	before := parseTags(meta.Tags)
	updated := mergeTags(meta.Tags, add, remove)
	after := parseTags(updated)

	added := diffTags(after, before)
	removed := diffTags(before, after)

	if updated != meta.Tags {
		meta.Tags = updated
		meta.Modified = c.Now().UTC()
		if err := storage.UpdateMetadata(ctx, c.DB, meta); err != nil {
			return nil, nil, err
		}
	}

	return added, removed, nil
}

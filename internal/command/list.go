package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	flag "github.com/spf13/pflag"
	"golang.org/x/term"

	"github.com/llywelwyn/wow/internal/model"
	"github.com/llywelwyn/wow/internal/storage"
	"github.com/llywelwyn/wow/internal/ui"
)

// ListCommand prints snippet metadata.
type ListCommand struct {
	DB     *sql.DB
	Output io.Writer
}

type listViewOptions struct {
	WithTags  bool
	WithDates bool
	WithDesc  bool
	WithType  bool
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
	var withTags *bool = fs.BoolP("with-tags", "t", false, "include tags")
	var withDates *bool = fs.BoolP("with-dates", "D", false, "include created/updated dates")
	var withDesc *bool = fs.BoolP("with-desc", "d", false, "include descriptions")
	var withType *bool = fs.BoolP("with-type", "T", false, "include snippet type")
	var all *bool = fs.BoolP("all", "a", false, "include all metadata (tags, type, dates, description)")
	var verbose *bool = fs.BoolP("verbose", "v", false, "alias for --all")
	var help *bool = fs.BoolP("help", "h", false, "display help")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *help {
		fmt.Fprintln(c.Output, `Usage:
  wow list [--plain[=delimiter]] [--with-tags] [--with-type]
          [--with-desc] [--with-dates] [--all/--verbose]

  Wow! Lists metadata for all the snippets you've got saved.

  List output is totally modular and can be adjusted with flags so
  only what you care about gets displayed. Running "wow list" with
  no flags shows a simple list of snippet keys.

  You can add fields one-by-one with flags or call list with --all
  to see all available information about each entry.

  Importantly: the view you see when you run "wow list" is pretty.
  Whenever you pipe the output somewhere else, the --plain flag is
  forcibly enabled to make parsing the data easier.

  --plain output is tab-delimited by default, but the flag accepts
  any string as an override if you prefer a different format.`)
		fmt.Fprintln(c.Output)
		fs.PrintDefaults()
		return nil
	}

	ctx := context.Background()
	entries, err := storage.ListMetadata(ctx, c.DB)
	if err != nil {
		return err
	}

	useAll := *all || *verbose
	opts := listViewOptions{
		WithTags:  *withTags || useAll,
		WithDates: *withDates || useAll,
		WithDesc:  *withDesc || useAll,
		WithType:  *withType || useAll,
	}

	if *plain != "" || !writerIsTerminal(c.Output) {
		delimiter := *plain
		if delimiter == "" {
			delimiter = "\t"
		}
		return renderPlainList(c.Output, entries, opts, delimiter)
	}
	return renderStyledList(c.Output, entries, opts)
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

func renderPlainList(w io.Writer, entries []model.Metadata, opts listViewOptions, delimiter string) error {
	for _, meta := range entries {
		fields := []string{meta.Key}
		if opts.WithType {
			fields = append(fields, meta.Type)
		}
		if opts.WithTags {
			fields = append(fields, plainTagList(meta.Tags))
		}
		if opts.WithDates {
			created := meta.Created.UTC().Format(time.DateOnly)
			modified := ""
			if !meta.Modified.Equal(meta.Created) {
				modified = meta.Modified.UTC().Format(time.DateOnly)
			}
			fields = append(fields, created, modified)
		}
		if opts.WithDesc {
			fields = append(fields, meta.Description)
		}

		if _, err := fmt.Fprintln(w, strings.Join(fields, delimiter)); err != nil {
			return err
		}
	}
	return nil
}

func renderStyledList(w io.Writer, entries []model.Metadata, opts listViewOptions) error {
	styles := ui.DefaultStyles()

	if len(entries) == 0 {
		_, err := fmt.Fprintln(w, styles.Empty.Render("(no snippets)"))
		return err
	}

	screenWidth := writerWidth(w)
	if screenWidth <= 0 {
		screenWidth = 80
	}
	contentWidth := max(screenWidth-4, 40)

	for _, meta := range entries {
		if _, err := fmt.Fprintln(w, renderStyledEntry(meta, contentWidth, styles, opts)); err != nil {
			return err
		}
	}
	return nil
}

func buildKeyLine(meta model.Metadata, styles ui.Styles, opts listViewOptions) string {
	if opts.WithType {
		icon := strings.TrimSpace(meta.TypeIcon())
		if icon != "" {
			return fmt.Sprintf("%s %s",
				styles.Icon.Render(icon),
				styles.Key.Render(meta.Key),
			)
		}
	}
	return styles.Key.Render(meta.Key)
}

func buildRootLine(meta model.Metadata, styles ui.Styles, wrap lipgloss.Style, opts listViewOptions) string {
	base := buildKeyLine(meta, styles, opts)
	if opts.WithTags {
		if tags := styledTagList(meta.Tags, styles); tags != "" {
			base = fmt.Sprintf("%s %s", base, tags)
		}
	}
	return wrap.Render(base)
}

func writerWidth(w io.Writer) int {
	type fdWriter interface {
		io.Writer
		Fd() uintptr
	}
	if f, ok := w.(fdWriter); ok {
		if wd, _, err := term.GetSize(int(f.Fd())); err == nil && wd > 0 {
			return wd
		}
	}
	return 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func renderStyledEntry(meta model.Metadata, width int, styles ui.Styles, opts listViewOptions) string {
	rootWidth := max(width, 40)
	childWidth := max(width-4, 20)

	rootWrap := lipgloss.NewStyle().
		Width(rootWidth).
		MaxWidth(rootWidth)
	childWrap := lipgloss.NewStyle().
		Width(childWidth).
		MaxWidth(childWidth)

	rootLine := buildRootLine(meta, styles, rootWrap, opts)

	t := tree.Root(rootLine).
		Enumerator(compactEnumerator).
		EnumeratorStyle(styles.Subtle).
		Indenter(compactIndenter).
		RootStyle(lipgloss.NewStyle()).
		ItemStyle(lipgloss.NewStyle())

	if opts.WithDates {
		if dateLine := buildDateLine(meta, styles); dateLine != "" {
			t.Child(childWrap.Render(dateLine))
		}
	}

	if opts.WithDesc {
		if desc := strings.TrimSpace(meta.Description); desc != "" {
			t.Child(childWrap.Render(styles.Subtle.Render(desc)))
		}
	}

	return t.String()
}

func buildDateLine(meta model.Metadata, styles ui.Styles) string {
	created := styles.Subtle.Render(meta.Created.UTC().Format(time.DateOnly))
	components := []string{
		fmt.Sprintf("%s %s", styles.Label.Render("created"), created),
	}
	if !meta.Modified.Equal(meta.Created) {
		modified := styles.Subtle.Render(meta.Modified.UTC().Format(time.DateOnly))
		components = append(components, fmt.Sprintf("%s %s", styles.Label.Render("last updated"), modified))
	}
	return strings.Join(components, "  ")
}

func styledTagList(raw string, styles ui.Styles) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	split := strings.Split(raw, ",")
	formatted := make([]string, 0, len(split))
	for _, part := range split {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		formatted = append(formatted, styles.Tag.Render("@"+part))
	}
	return strings.Join(formatted, " ")
}

func plainTagList(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	parts := strings.Split(raw, ",")
	formatted := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		formatted = append(formatted, "@"+part)
	}
	return strings.Join(formatted, " ")
}

func compactEnumerator(children tree.Children, index int) string {
	if children.Length()-1 == index {
		return "└─ "
	}
	return "├─ "
}

func compactIndenter(children tree.Children, index int) string {
	if children.Length()-1 == index {
		return "   "
	}
	return "│  "
}

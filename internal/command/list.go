package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"golang.org/x/term"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/model"
	"github.com/llywelwyn/wow/internal/storage"
	"github.com/llywelwyn/wow/internal/ui"
)

type ListCmd struct {
	PreviewLines  int     `short:"n" default:"1" name:"Preview" help:"Number of lines to preview from snippets."`
	Plain         bool    `help:"Format as a plain table of tab-separated values."`
	Tags          bool    `short:"t" help:"Display tags."`
	Date          bool    `short:"D" negatable:"" default:"true" help:"Display creation and last-modified dates."`
	Desc          bool    `short:"d" help:"Display description."`
	Type          bool    `short:"T" help:"Display type."`
	Verbose       bool    `short:"v" help:"Display all metadata."`
	Limit         int     `short:"l" default:"50" help:"Number of snippets per page."`
	Page          int     `short:"p" default:"1" help:"Page number."`
	All           bool    `short:"a" help:"Disable pagination and display all snippets."`
	Header        bool    `short:"H" negatable:"" default:"true" help:"Display header row."`
	TotalPages    int     `kong:"-"`
	TotalListings int     `kong:"-"`
	Columns       Columns `kong:"-"`
	BaseDir       string  `kong:"-"`
}

type Columns struct {
	Column map[string]Column
	Order  []string
}

func (c *Columns) SetWidths(list []model.Metadata) {
	// Init with blank header lengths.
	for name, col := range c.Column {
		col.Width = len(col.Header)
		c.Column[name] = col
	}
	// Get max widths from entries.
	for _, entry := range list {
		if c.Column["Date"].Shown {
			col := c.Column["Date"]
			col.Width = max(col.Width, len(entry.DateStr()))
			c.Column["Date"] = col
		}
		if c.Column["Type"].Shown {
			col := c.Column["Type"]
			col.Width = max(col.Width, len(entry.TypeStr()))
			c.Column["Type"] = col
		}
		if c.Column["Tags"].Shown {
			col := c.Column["Tags"]
			col.Width = max(col.Width, len(entry.TagsStr()))
			c.Column["Tags"] = col
		}
		if c.Column["Name"].Shown {
			col := c.Column["Name"]
			col.Width = max(col.Width, len(entry.NameStr()))
			c.Column["Name"] = col
		}
		if c.Column["Desc"].Shown {
			col := c.Column["Desc"]
			col.Width = max(col.Width, len(entry.DescStr()))
			c.Column["Desc"] = col
		}
	}
}

type Column struct {
	Header string
	Width  int
	Shown  bool
}

func (c *ListCmd) Run(kong *kong.Context, cfg Config) error {
	if c.Limit < 0 {
		return errors.New("limit must be >= 0")
	}

	if c.Page < 1 {
		return errors.New("page must be >= 1")
	}

	if c.All {
		c.Limit = 0
	}

	if c.Limit == 0 {
		c.Page = 1
	}

	c.Columns = Columns{
		Column: map[string]Column{
			"Date": {
				Header: "Date Modified",
				Width:  0,
				Shown:  c.Date || c.Verbose},
			"Type": {
				Header: "Type",
				Width:  0, Shown: c.Type || c.Verbose},
			"Name": {
				Header: "Name",
				Width:  0,
				Shown:  true},
			"Tags": {
				Header: "Tags",
				Width:  0,
				Shown:  c.Tags || c.Verbose},
			"Desc": {
				Header: "Desc",
				Width:  0,
				Shown:  c.Desc || c.Verbose},
		},
		Order: []string{"Date", "Type", "Name", "Tags", "Desc"},
	}

	c.BaseDir = cfg.BaseDir

	ctx := context.Background()

	listings, err := storage.ListMetadata(ctx, cfg.DB)
	if err != nil {
		return err
	}
	listings = c.paginate(listings)

	c.Columns.SetWidths(listings)

	c.print(cfg.Output, listings)

	return nil
}

func (c *ListCmd) paginate(listings []model.Metadata) []model.Metadata {
	// If no limit or all listings fit on one page, return all.
	c.TotalListings = len(listings)
	c.TotalPages = (c.TotalListings + c.Limit - 1) / c.Limit
	if c.Limit <= 0 || c.TotalListings <= c.Limit {
		return listings
	}

	// Clamp page number to valid range.
	c.Page = min(c.Page, c.TotalPages)
	c.Page = max(c.Page, 1)

	// Get our start index, return empty if past c.TotalListings.
	start := (c.Page - 1) * c.Limit
	if start >= c.TotalListings {
		return listings[:0]
	}

	// Get our end index, clamped to c.TotalListings.
	end := start + c.Limit
	if end > c.TotalListings {
		end = c.TotalListings
	}

	return listings[start:end]
}

func (c *ListCmd) print(w io.Writer, listings []model.Metadata) error {
	printFunc := c.pretty
	if c.Plain {
		printFunc = c.table
	}

	err := printFunc(w, listings)
	return err
}

func (c *ListCmd) buildHeader(style lipgloss.Style, delim string) ([]string, string, error) {
	var cols []string
	for _, name := range c.Columns.Order {
		if col, exists := c.Columns.Column[name]; name != "Desc" && exists && col.Shown {
			cols = append(cols, name)
		}
	}

	var headerString string
	if c.Header {
		var headers []string

		var lastCol string
		if len(cols) > 0 {
			lastCol = cols[len(cols)-1]
		}

		for _, colName := range cols {
			col := c.Columns.Column[colName]
			last := (colName == lastCol)

			if !last {
				style = style.Width(col.Width).AlignHorizontal(lipgloss.Left)
			}
			headers = append(headers, style.Render(col.Header))
		}
		headerString = strings.Join(headers, delim)
	}
	return cols, headerString, nil
}

func (c *ListCmd) table(w io.Writer, listings []model.Metadata) error {
	cols, headerString, err := c.buildHeader(lipgloss.NewStyle(), "\t")
	if err != nil {
		return err
	}

	if c.Header {
		fmt.Fprintln(w, headerString)
	}

	for _, listing := range listings {
		var fields []string
		for _, fieldName := range cols {
			switch fieldName {
			case "Name":
				fields = append(fields, listing.Key)
			case "Type":
				fields = append(fields, listing.Type)
			case "Tags":
				fields = append(fields, listing.Tags)
			case "Desc":
				fields = append(fields, listing.Description)
			case "Date":
				modified := listing.Modified.UTC().Format("02 Jan 15:04")
				fields = append(fields, modified)
			}
		}

		if _, err := fmt.Fprintln(w, strings.Join(fields, "\t")); err != nil {
			return err
		}
	}
	return nil
}

func (c *ListCmd) notShowingAll(first, final int) bool {
	return first > 1 || final < c.TotalListings
}

func (c *ListCmd) pretty(w io.Writer, listings []model.Metadata) error {
	styles := ui.GetStyles()

	// If we want to show the header (c.Header), and we're showing a subset of
	// our listings (e.g. we are showing a page, not everything), show page info.
	idxFirst := c.Limit*(c.Page-1) + 1
	idxFinal := min(c.Limit*c.Page, c.TotalListings)
	if c.Header && c.notShowingAll(idxFirst, idxFinal) {
		header := fmt.Sprintf("Page %d of %d (%d—%d/%d)",
			c.Page,
			c.TotalPages,
			idxFirst,
			idxFinal,
			c.TotalListings)

		fmt.Fprintln(w, styles.Body.Render(header))
	}

	// If we want headers, print them.
	if c.Header {
		cols, headerLine, err := c.buildHeader(styles.Underline, " ")
		if err != nil {
			return err
		}
		_ = cols
		if headerLine != "" {
			fmt.Fprintln(w, headerLine)
		}
	}

	if c.TotalListings == 0 {
		_, err := fmt.Fprintln(w, styles.Empty.Render("(no snippets)"))
		return err
	}

	// Default width to 80. Override with io.Writer width if possible.
	width := 80
	if wWidth := c.writerWidth(w); wWidth > 0 {
		width = wWidth
	}

	for _, listing := range listings {
		if _, err := fmt.Fprintln(w, c.prettyListing(listing, width, styles)); err != nil {
			return err
		}
		path, err := key.ResolvePath(c.BaseDir, listing.Key)
		if err != nil {
			return err
		}
		cmd := exec.Command("head", "-n", fmt.Sprint(c.PreviewLines), path)
		cmd.Stdout = w
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run head on %s: %w", listing.Key, err)
		}
	}

	return nil
}

func (c *ListCmd) prettyListing(listing model.Metadata, maxWidth int, styles ui.Styles) string {
	var lines []string
	var root []string
	var cols []string
	for _, name := range c.Columns.Order {
		if col, exists := c.Columns.Column[name]; exists && col.Shown {
			cols = append(cols, name)
		}
	}

	var lastRootCol string
	for i := len(cols) - 1; i >= 0; i-- {
		if cols[i] != "Desc" {
			lastRootCol = cols[i]
			break
		}
	}

	for _, name := range cols {
		if name == "Desc" {
			continue
		}

		col := c.Columns.Column[name]
		last := (name == lastRootCol)
		val := listing.Formatted(name)
		style := c.getColumnStyle(styles, name, col.Width, maxWidth, last)
		root = append(root, style.Render(val))
	}

	spacer := styles.Body.Width(1).Render(" ")
	var rootRender []string
	for i, part := range root {
		rootRender = append(rootRender, part)
		if i < len(root)-1 {
			rootRender = append(rootRender, spacer)
		}
	}
	lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, rootRender...))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (c *ListCmd) getColumnStyle(styles ui.Styles, name string, width, maxWidth int, last bool) lipgloss.Style {
	var style lipgloss.Style
	switch name {
	case "Date":
		style = styles.Body
	case "Type":
		style = styles.Primary
	case "Name":
		style = styles.Body
	case "Tags":
		style = styles.Tag
	case "Desc":
		style = styles.Muted
	}
	if !last {
		style = style.Width(width)
	} else if name == "Desc" {
		style = style.MaxWidth(maxWidth)
	}
	return style
}

func (c *ListCmd) writerWidth(w io.Writer) int {
	type fdWriter interface {
		io.Writer
		Fd() uintptr
	}

	f, ok := w.(fdWriter)
	if !ok {
		return 0
	}

	if width, _, err := term.GetSize(int(f.Fd())); err == nil && width > 0 {
		return width
	}

	return 0
}

type StringBuilder []string

func (s StringBuilder) Prefix(v string, ok bool) StringBuilder {
	if !ok {
		return s
	}
	return append(StringBuilder{v}, s...)
}

func (s StringBuilder) Suffix(v string, ok bool) StringBuilder {
	if !ok {
		return s
	}
	return append(s, v)
}

func (s StringBuilder) Build(sep string) string {
	return strings.Join(s, sep)
}

func (c *ListCmd) enumerator(children tree.Children, index int) string {
	if children.Length()-1 == index {
		return "└─ "
	}
	return "├─ "
}

func (c *ListCmd) indenter(children tree.Children, index int) string {
	if children.Length()-1 == index {
		return "   "
	}
	return "│  "
}

package command

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/ktr0731/go-fuzzyfinder"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/storage"
)

type GetCmd struct {
	Desc []string `arg:"" optional:""`
}

func (c *GetCmd) Run(kong *kong.Context, cfg Config) error {
	query := strings.Join(c.Desc, " ")

	if len(c.Desc) > 0 {
		// Try direct ID match.
		path, err := key.ResolvePath(cfg.BaseDir, c.Desc[0])
		if err == nil {
			if _, readErr := storage.Read(path); readErr == nil {
				return c.tryOutput(path, cfg.Output)
			}
		}
	}

	ctx := context.Background()
	metadata, err := storage.ListMetadata(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("list metadata: %w", err)
	}

	idx, err := fuzzyfinder.Find(
		metadata,
		func(i int) string {
			m := metadata[i]
			if m.Tags != "" {
				return fmt.Sprintf("%s [%s]", m.Description, m.Tags)
			}
			return m.Description
		},
		fuzzyfinder.WithQuery(query),
	)
	if err != nil {
		return fmt.Errorf("fuzzy find snippet: %w", err)
	}

	selected := metadata[idx].Key
	path, _ := key.ResolvePath(cfg.BaseDir, selected)
	return c.tryOutput(path, cfg.Output)
}

func (c *GetCmd) tryOutput(path string, w io.Writer) error {
	data, err := storage.Read(path)
	if err != nil {
		return fmt.Errorf("read snippet: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write snippet to output: %w", err)
	}
	return nil
}

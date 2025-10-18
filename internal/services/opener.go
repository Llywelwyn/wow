package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/storage"
)

// OpenOptions controls how a snippet should be opened.
type OpenOptions struct {
	UsePager bool
}

// Opener launches external programs to view snippets.
type Opener struct {
	BaseDir   string
	DB        *sql.DB
	OpenFunc  func(context.Context, string) error
	PagerFunc func(context.Context, string) error
}

// Open opens the snippet referred to by key with the configured program.
func (o *Opener) Open(ctx context.Context, rawKey string, opts OpenOptions) error {
	if o.DB == nil {
		return errors.New("opener misconfigured: db required")
	}
	if o.OpenFunc == nil {
		return errors.New("opener misconfigured: open func required")
	}
	if o.PagerFunc == nil {
		return errors.New("opener misconfigured: pager func required")
	}

	normalized, err := key.Normalize(rawKey)
	if err != nil {
		return err
	}

	meta, err := storage.GetMetadata(ctx, o.DB, normalized)
	if err != nil {
		return err
	}

	path, err := key.ResolvePath(o.BaseDir, normalized)
	if err != nil {
		return err
	}

	if opts.UsePager {
		return o.PagerFunc(ctx, path)
	}

	if meta.Type == "url" {
		data, err := storage.Read(path)
		if err != nil {
			return err
		}
		target := firstNonEmptyLine(data)
		if target == "" {
			target = string(data)
		}
		target = strings.TrimSpace(target)
		if target == "" {
			target = path
		}
		return o.OpenFunc(ctx, target)
	}

	return o.OpenFunc(ctx, path)
}

func firstNonEmptyLine(data []byte) string {
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		if trimmed := strings.TrimSpace(l); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

package command

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/llywelwyn/pda/internal/config"
	"github.com/llywelwyn/pda/internal/key"
	"github.com/llywelwyn/pda/internal/storage"
)

type EditCmd struct {
	Key string `arg:"" name:"key" help:"Snippet key."`
}

func (c *EditCmd) Run(kong *kong.Context, cfg config.Config) error {
	ctx := context.Background()
	// Resolve path and ensure it exists in the DB.
	normalizedKey, err := key.Normalize(c.Key)
	if err != nil {
		fmt.Println(err)
	}
	path, err := key.ResolvePath(cfg.BaseDir, normalizedKey)
	if err != nil {
		fmt.Println(err)
	}
	metadata, err := storage.GetMetadata(ctx, cfg.DB, normalizedKey)
	if err != nil {
		fmt.Println(err)
	}

	// Before -> Edit -> After
	before, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
	}
	if err := cfg.Editor(ctx, path); err != nil {
		fmt.Println(err)
	}
	after, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
	}

	// If there were changes, update Modified time and re-detect type.
	changed := after.ModTime() != before.ModTime() || after.Size() != before.Size()
	if !changed {
		return nil
	}
	data, err := storage.Read(path)
	if err != nil {
		fmt.Println(err)
	}
	metadata.Type = detectSnippetType(data)
	metadata.Modified = cfg.Clock().UTC()
	if err := storage.UpdateMetadata(ctx, cfg.DB, metadata); err != nil {
		fmt.Println(err)
	}

	return nil
}

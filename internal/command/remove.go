package command

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/llywelwyn/pda/internal/config"
	"github.com/llywelwyn/pda/internal/key"
	"github.com/llywelwyn/pda/internal/storage"
)

type RemoveCmd struct {
	Key string `arg:"" name:"key" help:"Snippet key."`
}

func (c *RemoveCmd) Run(ctx *kong.Context, cfg config.Config) error {
	normalizedKey, err := key.Normalize(c.Key)
	if err != nil {
		fmt.Fprintln(cfg.Output, err)
	}
	path, err := key.ResolvePath(cfg.BaseDir, normalizedKey)
	if err != nil {
		fmt.Fprintln(cfg.Output, err)
	}
	if err := storage.Delete(path); err != nil {
		fmt.Fprintln(cfg.Output, err)
	}
	if err := storage.DeleteMetadata(context.Background(), cfg.DB, normalizedKey); err != nil {
		fmt.Fprintln(cfg.Output, err)
	}
	return nil
}

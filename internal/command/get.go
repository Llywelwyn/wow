package command

import (
	"fmt"
	"io"

	"github.com/alecthomas/kong"
	"github.com/llywelwyn/wow/internal/config"
	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/storage"
)

type GetCmd struct {
	Desc []string `arg:"" optional:""`
}

func (c *GetCmd) Run(kong *kong.Context, cfg config.Config) error {
	if len(c.Desc) == 0 {
		return fmt.Errorf("no snippet specified")
	}

	if len(c.Desc) == 1 {
		tryId := c.Desc[0]
		path, err := key.ResolvePath(cfg.BaseDir, tryId)
		if err == nil {
			data, readErr := storage.Read(path)
			if readErr == nil {
				return c.writeOutput(data, cfg.Output)
			}
		}
	}

	return nil
}

func (c *GetCmd) writeOutput(data []byte, w io.Writer) error {
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write snippet to output: %w", err)
	}
	return nil
}

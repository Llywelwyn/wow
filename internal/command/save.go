package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/llywelwyn/pda/internal/config"
	"github.com/llywelwyn/pda/internal/key"
	"github.com/llywelwyn/pda/internal/model"
	"github.com/llywelwyn/pda/internal/storage"
	"github.com/llywelwyn/pda/internal/ui"
)

type SaveCmd struct {
	Desc []string `arg:"" optional:""`
	Tag  []string `help:"Comma-separated tags to add."`
}

func (c *SaveCmd) Run(kong *kong.Context, cfg config.Config) error {
	ctx := context.Background()

	payload, err := io.ReadAll(cfg.Input)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	if len(payload) == 0 {
		return errors.New("snippet content is empty")
	}

	now := cfg.Clock().UTC()
	snippetType := detectSnippetType(payload)

	nextId, err := cfg.NextId()
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	id := strconv.Itoa(nextId)

	path, err := key.ResolvePath(cfg.BaseDir, id)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	exists, err := storage.Exists(path)
	if err != nil {
		return fmt.Errorf("check snippet exists: %w", err)
	}
	if exists {
		return errors.New("snippet with that ID already exists")
	}

	if err := storage.Save(path, bytes.NewReader(payload)); err != nil {
		return fmt.Errorf("save snippt: %w", err)
	}

	metadata := model.Metadata{
		Key:         id,
		Type:        snippetType,
		Modified:    now,
		Description: strings.Join(c.Desc, " "),
		Tags:        strings.Join(c.Tag, ","),
	}

	if err := storage.InsertMetadata(ctx, cfg.DB, metadata); err != nil {
		_ = storage.Delete(path)
		if errors.Is(err, storage.ErrMetadataDuplicate) {
			return errors.New("snippet with that ID already exists")
		}
		return fmt.Errorf("insert metadata: %w", err)
	}

	fmt.Fprintln(cfg.Output, c.makeSaveOutput(metadata))

	return nil
}

func (c *SaveCmd) makeSaveOutput(metadata model.Metadata) string {
	styles := ui.GetStyles()
	desc := "snippet"
	if metadata.Description != "" {
		desc = fmt.Sprintf("'%s'", metadata.Description)
	}
	res := fmt.Sprintf("Saved %s with ID: %s", styles.Primary.Render(desc), styles.Key.Render(metadata.Key))
	return res
}

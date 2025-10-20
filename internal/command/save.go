package command

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/model"
	"github.com/llywelwyn/wow/internal/storage"
)

// SaveCommand persists snippet content read from stdin and prints the resolved key.
type SaveCommand struct {
	BaseDir string
	DB      *sql.DB
	Now     func() time.Time
	Input   io.Reader
	Output  io.Writer
}

// ErrSnippetExists indicates a save attempted to overwrite an existing snippet.
var ErrSnippetExists = errors.New("snippet already exists")

// NewSaveCommand constructs a SaveCommand using default dependencies from cfg.
func NewSaveCommand(cfg Config) *SaveCommand {
	return &SaveCommand{
		BaseDir: cfg.BaseDir,
		DB:      cfg.DB,
		Now:     cfg.clock(),
		Input:   cfg.reader(),
		Output:  cfg.writer(),
	}
}

// Name returns the command keyword for explicit invocation.
func (c *SaveCommand) Name() string {
	return "save"
}

// Execute saves the snippet using the provided arguments.
func (c *SaveCommand) Execute(args []string) error {
	if c.DB == nil || c.Now == nil || c.Input == nil || c.Output == nil || strings.TrimSpace(c.BaseDir) == "" {
		return errors.New("save command not fully configured")
	}

	tagArgs := extractTagArgs(args)
	args = tagArgs.Others

	fs := flag.NewFlagSet("save", flag.ContinueOnError)
	fs.SetOutput(c.Output)
	var tee *bool = fs.BoolP("tee", "T", false, "print stdin back out, rather than the key")
	var desc *string = fs.StringP("desc", "d", "", "description")
	var tags *string = fs.StringP("tag", "t", "", "comma-separated tags, e.g. one,two")
	var help *bool = fs.BoolP("help", "h", false, "display help")

	var keyArg string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		keyArg = args[0]
		args = args[1:]
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *help {
		fmt.Fprintln(c.Output, `Usage:
  wow save [key] [--desc description] [--tag tags] [@tag ...] < snippet`)
		fs.PrintDefaults()
		return nil
	}

	addTags := append(splitTags(*tags), tagArgs.Add...)

	resolvedKey, contents, err := c.saveSnippet(context.Background(), saveRequest{
		Key:         keyArg,
		Description: *desc,
		Tags:        addTags,
	})
	if err != nil {
		return err
	}

	output := resolvedKey
	if *tee {
		output = string(contents)
	}
	if _, err := fmt.Fprintln(c.Output, output); err != nil {
		return fmt.Errorf("write key to output: %w", err)
	}
	return nil
}

type saveRequest struct {
	Key         string
	Description string
	Tags        []string
}

func (c *SaveCommand) saveSnippet(ctx context.Context, req saveRequest) (string, []byte, error) {
	payload, err := io.ReadAll(c.Input)
	if err != nil {
		return "", nil, fmt.Errorf("read input: %w", err)
	}
	if len(payload) == 0 {
		return "", nil, errors.New("snippet content is empty")
	}

	now := c.Now().UTC()
	contentType := detectSnippetType(payload)

	resolvedKey, err := c.resolveKey(req.Key, now)
	if err != nil {
		return "", nil, err
	}

	path, err := key.ResolvePath(c.BaseDir, resolvedKey)
	if err != nil {
		return "", nil, err
	}

	exists, err := storage.Exists(path)
	if err != nil {
		return "", nil, err
	}
	if exists {
		return "", nil, ErrSnippetExists
	}

	if err := storage.Save(path, bytes.NewReader(payload)); err != nil {
		return "", nil, err
	}

	meta := model.Metadata{
		Key:         resolvedKey,
		Type:        contentType,
		Created:     now,
		Modified:    now,
		Description: req.Description,
		Tags:        mergeTags("", req.Tags, nil),
	}

	if err := storage.InsertMetadata(ctx, c.DB, meta); err != nil {
		_ = storage.Delete(path)
		if errors.Is(err, storage.ErrMetadataDuplicate) {
			return "", nil, ErrSnippetExists
		}
		return "", nil, err
	}

	return resolvedKey, payload, nil
}

func (c *SaveCommand) resolveKey(rawKey string, now time.Time) (string, error) {
	if strings.TrimSpace(rawKey) != "" {
		return rawKey, nil
	}

	cb := func(candidate string) (bool, error) {
		path, err := key.ResolvePath(c.BaseDir, candidate)
		if err != nil {
			return false, err
		}
		return storage.Exists(path)
	}

	autoKey, err := key.GenerateAuto(now, cb)
	if err != nil {
		return "", err
	}
	return autoKey, nil
}

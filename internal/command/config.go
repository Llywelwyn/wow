package command

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"os"
	"time"
)

// Config captures the common environment used to construct default commands.
type Config struct {
	BaseDir string
	DB      *sql.DB
	Input   io.Reader
	Output  io.Writer
	Clock   func() time.Time
	Editor  func(context.Context, string) error
	Opener  func(context.Context, string) error
	Pager   func(context.Context, string) error
}

func (c Config) reader() io.Reader {
	if c.Input != nil {
		return c.Input
	}
	return os.Stdin
}

func (c Config) writer() io.Writer {
	if c.Output != nil {
		return c.Output
	}
	return os.Stdout
}

func (c Config) clock() func() time.Time {
	if c.Clock != nil {
		return c.Clock
	}
	return time.Now
}

func (c Config) editor() func(context.Context, string) error {
	if c.Editor != nil {
		return c.Editor
	}
	return func(context.Context, string) error {
		return errors.New("editor opener not configured")
	}
}

func (c Config) opener() func(context.Context, string) error {
	if c.Opener != nil {
		return c.Opener
	}
	return func(context.Context, string) error {
		return errors.New("opener not configured")
	}
}

func (c Config) pager() func(context.Context, string) error {
	if c.Pager != nil {
		return c.Pager
	}
	return func(context.Context, string) error {
		return errors.New("pager not configured")
	}
}

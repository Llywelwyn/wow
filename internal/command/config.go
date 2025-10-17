package command

import (
	"context"
	"database/sql"
	"io"
	"os"
	"time"
)

// Config captures the common environment used to construct default commands.
type Config struct {
	BaseDir    string
	DB         *sql.DB
	Input      io.Reader
	Output     io.Writer
	Clock      func() time.Time
	EditorOpen func(context.Context, string) error
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

func (c Config) editorOpen() func(context.Context, string) error {
	return c.EditorOpen
}

package editor

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

// Command resolves the editor executable using WOW_EDITOR, EDITOR, or a nano fallback.
func Command() string {
	if cmd := strings.TrimSpace(os.Getenv("WOW_EDITOR")); cmd != "" {
		return cmd
	}
	if cmd := strings.TrimSpace(os.Getenv("EDITOR")); cmd != "" {
		return cmd
	}
	return "nano"
}

// Opener returns a function that launches the provided editor command.
func Opener(editor string) func(context.Context, string) error {
	return func(ctx context.Context, path string) error {
		cmd := exec.CommandContext(ctx, editor, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

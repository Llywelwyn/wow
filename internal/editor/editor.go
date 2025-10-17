package editor

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

// Command resolves the editor executable using WOW_EDITOR, EDITOR, or a nano fallback.
func GetEditorFromEnv() string {
	if cmd := strings.TrimSpace(os.Getenv("WOW_EDITOR")); cmd != "" {
		return cmd
	}
	if cmd := strings.TrimSpace(os.Getenv("EDITOR")); cmd != "" {
		return cmd
	}
	return "nano"
}

// OpenPath returns a function that launches the provided editor command.
func OpenPath(editor string) func(context.Context, string) error {
	return func(ctx context.Context, path string) error {
		parts := strings.Fields(strings.TrimSpace(editor))
		if len(parts) == 0 {
			panic("editor path is empty")
		}
		name, args := parts[0], parts[1:]
		cmd := exec.CommandContext(ctx, name, append(args, path)...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

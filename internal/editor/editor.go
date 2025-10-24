package editor

import (
	"os"
	"strings"
)

// GetEditorFromEnv resolves the editor executable using PDA_EDITOR, EDITOR, or a nano fallback.
func GetEditorFromEnv() string {
	if cmd := strings.TrimSpace(os.Getenv("PDA_EDITOR")); cmd != "" {
		return cmd
	}
	if cmd := strings.TrimSpace(os.Getenv("EDITOR")); cmd != "" {
		return cmd
	}
	return "nano"
}

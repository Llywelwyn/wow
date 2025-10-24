package opener

import (
	"os"
	"strings"
)

// GetOpenerFromEnv resolves the opener command from environment variables.
// It checks PDA_OPENER first and falls back to xdg-open.
func GetOpenerFromEnv() string {
	if cmd := strings.TrimSpace(os.Getenv("PDA_OPENER")); cmd != "" {
		return cmd
	}
	return "xdg-open"
}

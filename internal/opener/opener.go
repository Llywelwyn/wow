package opener

import (
	"os"
	"strings"
)

// GetOpenerFromEnv resolves the opener command from environment variables.
// It checks WOW_OPENER first and falls back to xdg-open.
func GetOpenerFromEnv() string {
	if cmd := strings.TrimSpace(os.Getenv("WOW_OPENER")); cmd != "" {
		return cmd
	}
	return "xdg-open"
}

package pager

import (
	"os"
	"strings"
)

// GetPagerFromEnv resolves the pager command from environment variables.
// It prefers WOW_PAGER, then PAGER, and finally defaults to less.
func GetPagerFromEnv() string {
	if cmd := strings.TrimSpace(os.Getenv("WOW_PAGER")); cmd != "" {
		return cmd
	}
	if cmd := strings.TrimSpace(os.Getenv("PAGER")); cmd != "" {
		return cmd
	}
	return "less"
}

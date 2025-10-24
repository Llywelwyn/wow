package opener

import "testing"

func TestGetOpenerFromEnvUsesOverride(t *testing.T) {
	t.Setenv("PDA_OPENER", "custom-open")
	if got := GetOpenerFromEnv(); got != "custom-open" {
		t.Fatalf("GetOpenerFromEnv() = %q, want %q", got, "custom-open")
	}
}

func TestGetOpenerFromEnvFallback(t *testing.T) {
	t.Setenv("PDA_OPENER", "")
	if got := GetOpenerFromEnv(); got != "xdg-open" {
		t.Fatalf("GetOpenerFromEnv() = %q, want xdg-open", got)
	}
}

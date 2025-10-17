package editor

import "testing"

func TestCommandPrefersWOWEditor(t *testing.T) {
	t.Setenv("WOW_EDITOR", "code")
	t.Setenv("EDITOR", "vim")

	if got := Command(); got != "code" {
		t.Fatalf("Command() = %q, want %q", got, "code")
	}
}

func TestCommandFallsBackToEditor(t *testing.T) {
	t.Setenv("WOW_EDITOR", "")
	t.Setenv("EDITOR", "vim")

	if got := Command(); got != "vim" {
		t.Fatalf("Command() = %q, want %q", got, "vim")
	}
}

func TestCommandDefault(t *testing.T) {
	t.Setenv("WOW_EDITOR", "")
	t.Setenv("EDITOR", "")

	if got := Command(); got != "nano" {
		t.Fatalf("Command() = %q, want %q", got, "nano")
	}
}

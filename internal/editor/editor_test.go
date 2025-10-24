package editor

import "testing"

func TestCommandPreferspdaEditor(t *testing.T) {
	t.Setenv("PDA_EDITOR", "code")
	t.Setenv("EDITOR", "vim")

	if got := GetEditorFromEnv(); got != "code" {
		t.Fatalf("GetEditorFromEnv() = %q, want %q", got, "code")
	}
}

func TestCommandFallsBackToEditor(t *testing.T) {
	t.Setenv("PDA_EDITOR", "")
	t.Setenv("EDITOR", "vim")

	if got := GetEditorFromEnv(); got != "vim" {
		t.Fatalf("GetEditorFromEnv() = %q, want %q", got, "vim")
	}
}

func TestCommandDefault(t *testing.T) {
	t.Setenv("PDA_EDITOR", "")
	t.Setenv("EDITOR", "")

	if got := GetEditorFromEnv(); got != "nano" {
		t.Fatalf("GetEditorFromEnv() = %q, want %q", got, "nano")
	}
}

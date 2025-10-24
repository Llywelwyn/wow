package config

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestLoadPreferspdaHome(t *testing.T) {
	want := filepath.Join(t.TempDir(), "pdahome")

	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("PDA_HOME", want)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseDir != want {
		t.Fatalf("BaseDir = %q, want %q", cfg.BaseDir, want)
	}
}

func TestLoadFallsBackToXDG(t *testing.T) {
	// Using generic XDG home, we expect /pda to be appended.
	xdg_data_home := filepath.Join(t.TempDir(), "pdahome")
	want := filepath.Join(xdg_data_home, "pda")

	t.Setenv("PDA_HOME", "")
	t.Setenv("XDG_DATA_HOME", xdg_data_home)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseDir != want {
		t.Fatalf("BaseDir = %q, want %q", cfg.BaseDir, want)
	}
}

func TestLoadFallsBackToHome(t *testing.T) {
	// Using home, we expect /.pda to be appended.
	home := t.TempDir()
	want := filepath.Join(home, ".pda")

	t.Setenv("PDA_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", home)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseDir != want {
		t.Fatalf("BaseDir = %q, want %q", cfg.BaseDir, want)
	}
}

func TestLoadExpandsTilde(t *testing.T) {
	home := t.TempDir()
	folder := "pda-data"
	want := filepath.Join(home, folder)

	t.Setenv("PDA_HOME", fmt.Sprintf("~/%s", folder))
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", home)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseDir != want {
		t.Fatalf("BaseDir = %q, want %q", cfg.BaseDir, want)
	}
}

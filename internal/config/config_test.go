package config

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestLoadPrefersWowHome(t *testing.T) {
	want := filepath.Join(t.TempDir(), "wowhome")

	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("WOW_HOME", want)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseDir != want {
		t.Fatalf("BaseDir = %q, want %q", cfg.BaseDir, want)
	}
}

func TestLoadFallsBackToXDG(t *testing.T) {
	// Using generic XDG home, we expect /wow to be appended.
	xdg_data_home := filepath.Join(t.TempDir(), "wowhome")
	want := filepath.Join(xdg_data_home, "wow")

	t.Setenv("WOW_HOME", "")
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
	// Using home, we expect /.wow to be appended.
	home := t.TempDir()
	want := filepath.Join(home, ".wow")

	t.Setenv("WOW_HOME", "")
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
	folder := "wow-data"
	want := filepath.Join(home, folder)

	t.Setenv("WOW_HOME", fmt.Sprintf("~/%s", folder))
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

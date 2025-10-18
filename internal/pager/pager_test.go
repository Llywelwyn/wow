package pager

import "testing"

func TestGetPagerFromEnvPrefersWowPager(t *testing.T) {
	t.Setenv("WOW_PAGER", "bat")
	t.Setenv("PAGER", "less")
	if got := GetPagerFromEnv(); got != "bat" {
		t.Fatalf("GetPagerFromEnv() = %q, want %q", got, "bat")
	}
}

func TestGetPagerFromEnvFallsBackToPager(t *testing.T) {
	t.Setenv("WOW_PAGER", "")
	t.Setenv("PAGER", "more")
	if got := GetPagerFromEnv(); got != "more" {
		t.Fatalf("GetPagerFromEnv() = %q, want %q", got, "more")
	}
}

func TestGetPagerFromEnvDefaults(t *testing.T) {
	t.Setenv("WOW_PAGER", "")
	t.Setenv("PAGER", "")
	if got := GetPagerFromEnv(); got != "less" {
		t.Fatalf("GetPagerFromEnv() = %q, want %q", got, "less")
	}
}

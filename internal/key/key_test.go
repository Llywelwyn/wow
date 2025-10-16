package key

import (
	"path/filepath"
	"testing"
)

func TestNormalizeAcceptsValidKeys(t *testing.T) {
	tests := []string{"foo", "foo/bar", "foo/bar/baz", "foo_/bar/baz-123"}
	for _, tc := range tests {
		got, err := Normalize(tc)
		if err != nil {
			t.Fatalf("Normalize(%q) error = %v", tc, err)
		}
		if got != tc {
			t.Fatalf("Normalize(%q) = %q, want %q", tc, got, tc)
		}
	}
}

func TestNormalizeRejectsLeadingSlash(t *testing.T) {
	tc := "/foo"
	if _, err := Normalize(tc); err != ErrAbsolute {
		t.Fatalf("Normalize(%q) = %v, expected ErrAbsolute", tc, err)
	}
}

func TestNormalizeRejectsDots(t *testing.T) {
	tests := []string{"..", "foo/..", ".", "foo/./bar"}
	for _, tc := range tests {
		if _, err := Normalize(tc); err == nil {
			t.Fatalf("Normalize(%q) expected error", tc)
		}
	}
}

func TestNormalizeRejectsInvalidCharacters(t *testing.T) {
	tests := []string{"foo bar", "foo🔥bar"}
	for _, tc := range tests {
		if _, err := Normalize(tc); err == nil {
			t.Fatalf("Normalize(%q) expected error", tc)
		}
	}
}

func TestResolveProducesSafePath(t *testing.T) {
	base := t.TempDir()
	tc := "auto/1700000000"

	have, err := Resolve(base, tc)
	if err != nil {
		t.Fatalf("Resolve(%q, %q) error = %v", base, tc, err)
	}

	want := filepath.Join(base, tc)
	if have != want {
		t.Fatalf("Resolve(%q, %q) = %q, want %q", base, tc, have, want)
	}
}

func TestResolveRejectsTraversal(t *testing.T) {
	tests := []string{"foo/../../bar", "../foo", ".."}
	base := t.TempDir()
	for _, tc := range tests {
		if _, err := Resolve(base, tc); err == nil {
			t.Fatalf("Resolve(%q, %q) expected error", base, tc)
		}
	}
}

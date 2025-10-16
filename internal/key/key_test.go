package key

import (
	"path/filepath"
	"testing"
)

func TestResolvePathProducesSafePath(t *testing.T) {
	base := t.TempDir()
	tc := "auto/1700000000"

	have, err := ResolvePath(base, tc)
	if err != nil {
		t.Fatalf("ResolvePath(%q, %q) error = %v", base, tc, err)
	}

	want := filepath.Join(base, tc)
	if have != want {
		t.Fatalf("ResolvePath(%q, %q) = %q, want %q", base, tc, have, want)
	}
}

func TestResolvePathRejectsTraversal(t *testing.T) {
	tests := []string{"foo/../../bar", "../foo", ".."}
	base := t.TempDir()
	for _, tc := range tests {
		if _, err := ResolvePath(base, tc); err == nil {
			t.Fatalf("ResolvePath(%q, %q) expected error", base, tc)
		}
	}
}

package command

import "testing"

func TestMergeTagsAdd(t *testing.T) {
	got := mergeTags("", []string{"Foo", "bar"}, nil)
	if got != "foo,bar" {
		t.Fatalf("mergeTags = %q, want foo,bar", got)
	}
}

func TestMergeTagsRemove(t *testing.T) {
	got := mergeTags("foo,bar,baz", nil, []string{"BAR"})
	if got != "foo,baz" {
		t.Fatalf("mergeTags = %q, want foo,baz", got)
	}
}

func TestMergeTagsAddAndRemove(t *testing.T) {
	got := mergeTags("foo", []string{"bar"}, []string{"foo"})
	if got != "bar" {
		t.Fatalf("mergeTags = %q, want bar", got)
	}
}

func TestMergeTagsDedupes(t *testing.T) {
	got := mergeTags("foo", []string{"foo", "Foo"}, nil)
	if got != "foo" {
		t.Fatalf("mergeTags = %q, want foo", got)
	}
}

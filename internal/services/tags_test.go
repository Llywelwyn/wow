package services

import "testing"

func TestMergeTagsAdd(t *testing.T) {
	got := MergeTags("", []string{"Foo", "bar"}, nil)
	if got != "foo,bar" {
		t.Fatalf("MergeTags = %q, want foo,bar", got)
	}
}

func TestMergeTagsRemove(t *testing.T) {
	got := MergeTags("foo,bar,baz", nil, []string{"BAR"})
	if got != "foo,baz" {
		t.Fatalf("MergeTags = %q, want foo,baz", got)
	}
}

func TestMergeTagsAddAndRemove(t *testing.T) {
	got := MergeTags("foo", []string{"bar"}, []string{"foo"})
	if got != "bar" {
		t.Fatalf("MergeTags = %q, want bar", got)
	}
}

func TestMergeTagsDedupes(t *testing.T) {
	got := MergeTags("foo", []string{"foo", "Foo"}, nil)
	if got != "foo" {
		t.Fatalf("MergeTags = %q, want foo", got)
	}
}

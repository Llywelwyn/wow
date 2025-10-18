package command

import (
	"context"
	"errors"
	"testing"

	"github.com/llywelwyn/wow/internal/services"
)

type stubOpenService struct {
	called bool
	key    string
	opts   services.OpenOptions
	err    error
}

func (s *stubOpenService) Open(ctx context.Context, key string, opts services.OpenOptions) error {
	s.called = true
	s.key = key
	s.opts = opts
	return s.err
}

func TestOpenCommandName(t *testing.T) {
	cmd := &OpenCommand{}
	if cmd.Name() != "open" {
		t.Fatalf("Name = %q, want open", cmd.Name())
	}
}

func TestOpenCommandRequiresOpener(t *testing.T) {
	cmd := &OpenCommand{}
	if err := cmd.Execute([]string{"key"}); err == nil {
		t.Fatalf("expected configuration error")
	}
}

func TestOpenCommandRequiresKey(t *testing.T) {
	cmd := &OpenCommand{Opener: &stubOpenService{}}
	if err := cmd.Execute(nil); err == nil {
		t.Fatalf("expected error for missing key")
	}
}

func TestOpenCommandRequiresSingleKey(t *testing.T) {
	cmd := &OpenCommand{Opener: &stubOpenService{}}
	if err := cmd.Execute([]string{"one", "two"}); err == nil {
		t.Fatalf("expected error for too many args")
	}
}

func TestOpenCommandCallsService(t *testing.T) {
	stub := &stubOpenService{}
	cmd := &OpenCommand{Opener: stub}
	if err := cmd.Execute([]string{"key"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if !stub.called || stub.key != "key" {
		t.Fatalf("expected service called with key")
	}
	if stub.opts.UsePager {
		t.Fatalf("expected pager false")
	}
}

func TestOpenCommandPagerFlag(t *testing.T) {
	stub := &stubOpenService{}
	cmd := &OpenCommand{Opener: stub}
	if err := cmd.Execute([]string{"--pager", "key"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if !stub.opts.UsePager {
		t.Fatalf("expected UsePager true")
	}
}

func TestOpenCommandPropagatesError(t *testing.T) {
	stub := &stubOpenService{err: errors.New("boom")}
	cmd := &OpenCommand{Opener: stub}
	if err := cmd.Execute([]string{"key"}); err != stub.err {
		t.Fatalf("expected %v, got %v", stub.err, err)
	}
}

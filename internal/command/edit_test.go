package command

import (
	"context"
	"errors"
	"testing"

	"github.com/llywelwyn/wow/internal/model"
)

func TestEditCommandName(t *testing.T) {
	cmd := &EditCommand{}
	if cmd.Name() != "edit" {
		t.Fatalf("Name = %q, want edit", cmd.Name())
	}
}

func TestEditCommandRequiresEditor(t *testing.T) {
	cmd := &EditCommand{}
	if err := cmd.Execute([]string{"key"}); err == nil {
		t.Fatalf("expected error when editor is nil")
	}
}

func TestEditCommandRequiresSingleKey(t *testing.T) {
	editor := &stubEditor{}
	cmd := &EditCommand{Editor: editor}

	if err := cmd.Execute(nil); err == nil {
		t.Fatalf("expected error for missing key")
	}

	if err := cmd.Execute([]string{"one", "two"}); err == nil {
		t.Fatalf("expected error for too many args")
	}
}

func TestEditCommandCallsEditor(t *testing.T) {
	editor := &stubEditor{
		edit: func(ctx context.Context, key string) (model.Metadata, error) {
			if ctx == nil {
				t.Fatalf("expected context")
			}
			if key != "key" {
				t.Fatalf("expected key, got %q", key)
			}
			return model.Metadata{}, nil
		},
	}
	cmd := &EditCommand{Editor: editor}
	if err := cmd.Execute([]string{"key"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if !editor.called {
		t.Fatalf("expected editor.Edit to be called")
	}
}

func TestEditCommandPropagatesErrors(t *testing.T) {
	wantErr := errors.New("boom")
	editor := &stubEditor{
		edit: func(ctx context.Context, key string) (model.Metadata, error) {
			return model.Metadata{}, wantErr
		},
	}
	cmd := &EditCommand{Editor: editor}
	if err := cmd.Execute([]string{"key"}); err != wantErr {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

type stubEditor struct {
	edit   func(ctx context.Context, key string) (model.Metadata, error)
	called bool
}

func (s *stubEditor) Edit(ctx context.Context, key string) (model.Metadata, error) {
	s.called = true
	if s.edit == nil {
		return model.Metadata{}, nil
	}
	return s.edit(ctx, key)
}

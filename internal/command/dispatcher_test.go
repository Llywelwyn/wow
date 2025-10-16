package command

import (
	"errors"
	"testing"
)

type stubCommand struct {
	name   string
	called bool
	args   []string
	retErr error
}

func (s *stubCommand) Name() string { return s.name }
func (s *stubCommand) Execute(args []string) error {
	s.called = true
	s.args = args
	return s.retErr
}

func TestDispatchNilReturnsNotYetImplemented(t *testing.T) {
	d := NewDispatcher()
	err := d.Dispatch(nil)
	if !errors.Is(err, ErrNotYetImplemented) {
		t.Fatalf("expected ErrNotYetImplemented, got %v", err)
	}
}

func TestDispatcherUnknownCommand(t *testing.T) {
	d := NewDispatcher()
	err := d.Dispatch([]string{"missing"})
	if !errors.Is(err, ErrUnknownCommand) {
		t.Fatalf("expected ErrUnknownCommand, got %v", err)
	}
}

func TestDispatcherInvokesRegisteredCommand(t *testing.T) {
	cmd := &stubCommand{name: "ls"}
	args := []string{"ls", "-v"}
	d := NewDispatcher()
	d.Register(cmd)

	if err := d.Dispatch(args); err != nil {
		t.Fatalf("Dispatch(%q) error = %v", cmd.args, err)
	}
	if !cmd.called {
		t.Fatalf("expected command to be called")
	}
	if len(cmd.args) != 1 || cmd.args[0] != "-v" {
		t.Fatalf("expected forwarded args, got %v", cmd.args)
	}
}

func TestDispatcherLookup(t *testing.T) {
	cmd := &stubCommand{name: "save"}
	d := NewDispatcher()
	d.Register(cmd)

	found, ok := d.Lookup("save")
	if !ok {
		t.Fatalf("expected to find registered command")
	}
	if found != cmd {
		t.Fatalf("lookup returned unexpected command")
	}

	if _, ok := d.Lookup("missing"); ok {
		t.Fatalf("expected lookup to fail for missing command")
	}
}

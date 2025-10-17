package runner

import (
	"context"
	"errors"
	"os/exec"
	"testing"
)

func TestCommandEmpty(t *testing.T) {
    if _, err := Command(" \t "); err == nil {
        t.Fatalf("expected error for empty command")
    }
}

func TestCommandInvokesProcess(t *testing.T) {
	fn, err := Command("echo")
	if err != nil {
		t.Fatalf("Command error = %v", err)
	}

	// Execute with harmless argument; the command exists on POSIX.
	if err := fn(context.Background(), "runner"); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("unexpected error type: %v", err)
		}
	}
}

func TestRunReturnsErrorClosureForEmpty(t *testing.T) {
    fn := Run("  ")
    if err := fn(context.Background(), "target"); err == nil {
        t.Fatalf("expected error from invalid command")
    }
}

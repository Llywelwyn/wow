package runner

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
)

// Command returns a callable that executes the provided command string with the target argument.
func Command(command string) (func(context.Context, string) error, error) {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return nil, errors.New("command is empty")
	}

	parts := strings.Fields(trimmed)
	name := parts[0]
	baseArgs := append([]string(nil), parts[1:]...)

	return func(ctx context.Context, target string) error {
		args := append(append([]string(nil), baseArgs...), target)
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}, nil
}

// Run wraps Command and returns a function even when parsing fails.
// If the command string is invalid, the returned function always returns that error.
func Run(command string) func(context.Context, string) error {
	fn, err := Command(command)
	if err != nil {
		return func(context.Context, string) error { return err }
	}
	return fn
}

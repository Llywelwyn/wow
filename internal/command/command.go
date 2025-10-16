package command

import "errors"

// Command describes a runnable CLI command.
type Command interface {
	Name() string                // returns the command requested.
	Execute(args []string) error // executes command with *args.
}

var (
	ErrUnknownCommand    = errors.New("unknown command")
	ErrNotYetImplemented = errors.New("not yet implemented")
)

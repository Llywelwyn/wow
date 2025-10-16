package command

import "fmt"

// Dispatcher routes CLI args to registered commands.
type Dispatcher struct {
	registry map[string]Command
}

// NewDispatcher constructs an empty Dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		registry: make(map[string]Command),
	}
}

// Register makes the command available to the Dispatcher.
func (d *Dispatcher) Register(cmd Command) {
	if cmd == nil {
		panic("cannot register nil command")
	}
	name := cmd.Name()
	if name == "" {
		panic("command name must not be empty")
	}
	if _, exists := d.registry[name]; exists {
		panic("cannot register the same command twice")
	}
	d.registry[name] = cmd
}

// Dispatch selects a command based on args and invokes it.
func (d *Dispatcher) Dispatch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: default command pending", ErrNotYetImplemented)
	}
	name := args[0]
	cmd, ok := d.Lookup(name)
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownCommand, name)
	}
	if err := cmd.Execute(args[1:]); err != nil {
		return fmt.Errorf("command %q failed: %w", name, err)
	}
	return nil
}

// Lookup returns the command registered under the provided name.
func (d *Dispatcher) Lookup(name string) (Command, bool) {
	cmd, ok := d.registry[name]
	return cmd, ok
}

package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/llywelwyn/wow/internal/command"
	"github.com/llywelwyn/wow/internal/config"
	"github.com/llywelwyn/wow/internal/editor"
	"github.com/llywelwyn/wow/internal/runner"
	"github.com/llywelwyn/wow/internal/storage"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := storage.InitMetaDB(cfg.MetaDB)
	if err != nil {
		return err
	}
	defer db.Close()

	dispatcher := command.NewDispatcher()

	cmdCfg := command.Config{
		BaseDir: cfg.BaseDir,
		DB:      db,
		Input:   os.Stdin,
		Output:  os.Stdout,
		Clock:   time.Now,
		Editor:  runner.Run(editor.GetEditorFromEnv()),
	}

	saveCmd := command.NewSaveCommand(cmdCfg)
	getCmd := command.NewGetCommand(cmdCfg)
	editCmd := command.NewEditCommand(cmdCfg)
	listCmd := command.NewListCommand(cmdCfg)
	removeCmd := command.NewRemoveCommand(cmdCfg)

	dispatcher.Register(saveCmd)
	dispatcher.Register(getCmd)
	dispatcher.Register(editCmd)
	dispatcher.Register(listCmd, "ls")
	dispatcher.Register(removeCmd, "rm")

	// os.Args[0] is this script. Take the rest.
	args := os.Args[1:]

	piped, err := stdinHasData()
	if err != nil {
		return err
	}

	// If no args, check for stdin to save implicitly.
	// Otherwise just print usage.
	if len(args) == 0 {
		if piped {
			// echo "data" | wow
			// Implicit save with auto-generated key.
			return saveCmd.Execute(nil)
		}
		printUsage()
		return nil
	}

	// Check for explicit command in args[0]
	// and execute if a match is found.
	if cmd, ok := dispatcher.Lookup(args[0]); ok {
		// wow ls --verbose
		// wow edit <key>
		return cmd.Execute(args[1:])
	}

	// Implicit save or get based on stdin presence.
	if piped {
		// echo "func foo() {}" | wow go/foo
		// wow go/bar < bar.go
		return saveCmd.Execute(args)
	}
	return getCmd.Execute(args)
}

func stdinHasData() (bool, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		if errors.Is(err, os.ErrClosed) {
			return false, nil
		}
		return false, err
	}
	return info.Mode()&os.ModeCharDevice == 0, nil
}

func printUsage() {
	fmt.Fprintf(os.Stdout, `Usage:
  wow [key]                        Retrieve snippet when no stdin data
  wow [key] < file                 Save snippet with explicit key
  wow < file                       Save snippet with auto-generated key
  wow save [key]                   Explicit save
  wow get <key>                    Explicit get
  wow edit <key>                   Edit snippet in $WOW_EDITOR or $EDITOR
  wow list [--verbose] [--plain]   List saved snippets (alias: ls)
  wow remove <key>                 Remove snippet (alias: rm)
`)
}

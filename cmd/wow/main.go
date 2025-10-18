package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/llywelwyn/wow/internal/command"
	"github.com/llywelwyn/wow/internal/config"
	"github.com/llywelwyn/wow/internal/editor"
	"github.com/llywelwyn/wow/internal/opener"
	"github.com/llywelwyn/wow/internal/pager"
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
		Opener:  runner.Run(opener.GetOpenerFromEnv()),
		Pager:   runner.Run(pager.GetPagerFromEnv()),
	}

	saveCmd := command.NewSaveCommand(cmdCfg)
	getCmd := command.NewGetCommand(cmdCfg)
	editCmd := command.NewEditCommand(cmdCfg)
	openCmd := command.NewOpenCommand(cmdCfg)
	listCmd := command.NewListCommand(cmdCfg)
	removeCmd := command.NewRemoveCommand(cmdCfg)

	dispatcher.Register(saveCmd)
	dispatcher.Register(getCmd)
	dispatcher.Register(editCmd)
	dispatcher.Register(openCmd)
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

	if args[0] == "--help" || args[0] == "-h" {
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
  wow [key] [--tag string] [--untag string] [@tag] [-@tag]    	Retrieve snippet when no stdin data
  wow <key> [--tag string] [@tag] [--desc string] < file      	Save snippet with explicit key
  wow [--tag string] [@tag] [--desc string] < file            	Save snippet with auto-generated key
  wow save <key> [--tag string] [@tag] [--desc string] < file	Explicit save
  wow get <key> [--tag string] [--untag string] [@tag] [-@tag]	Explicit get
  wow open <key> [--pager]                                    	Open snippet or view in pager
  wow edit <key>                                              	Edit snippet in $WOW_EDITOR or $EDITOR
  wow list [--plain] [--quiet]                                	List saved snippets (alias: ls)
  wow remove <key>                                            	Remove snippet (alias: rm)

  Run any command with --help for more info on that specific command.

  Any flags can be written in shorthand (with a single dash, and their first letter).
  For example:
   wow list --plain --quiet 	-->	wow ls -pq
   wow <key> --tag tag1,tag2	-->	wow <key> -t=tag1,tag2


  When combining short flags, only the final flag can accept an argument. If you need
  to pass arguments to multiple flags, they must be written separately.
`)
}

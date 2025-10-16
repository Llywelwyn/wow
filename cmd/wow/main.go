package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/llywelwyn/wow/internal/command"
	"github.com/llywelwyn/wow/internal/config"
	"github.com/llywelwyn/wow/internal/core"
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

	saver := &core.Saver{
		BaseDir: cfg.BaseDir,
		DB:      db,
		Now:     time.Now,
	}

	saveCmd := &command.SaveCommand{
		Saver:  saver,
		Input:  os.Stdin,
		Output: os.Stdout,
	}
	getCmd := &command.GetCommand{
		BaseDir: cfg.BaseDir,
		Output:  os.Stdout,
	}

	dispatcher.Register(saveCmd)
	dispatcher.Register(getCmd)

	args := os.Args[1:]
	piped, err := stdinHasData()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		if piped {
			return saveCmd.Execute(nil)
		}
		printUsage()
		return nil
	}

	if cmd, ok := dispatcher.Lookup(args[0]); ok {
		return cmd.Execute(args[1:])
	}

	if piped {
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
  wow [key]            Retrieve snippet when no stdin data
  wow [key] < file     Save snippet with explicit key
  wow < file           Save snippet with auto-generated key

Commands:
  wow save [key]       Explicit save
  wow get <key>        Explicit get
`)
}

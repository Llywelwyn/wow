package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/llywelwyn/pda/internal/command"
	"github.com/llywelwyn/pda/internal/config"

	"github.com/alecthomas/kong"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

var pda struct {
	Save   command.SaveCmd   `cmd:"1" aliases:"s" help:"Save a snippet."`
	Get    command.GetCmd    `cmd:"1" aliases:"g" help:"Get a snippet."`
	Edit   command.EditCmd   `cmd:"1" aliases:"e" help:"Edit a snippet."`
	List   command.ListCmd   `cmd:"1" aliases:"l,ls" help:"List snippets."`
	Remove command.RemoveCmd `cmd:"1" aliases:"r,rm" help:"Remove a snippet."`
}

func root() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	defer cfg.DB.Close()

	description := `
 pda! a tool for code snippets

 ██     ██  ██████  ██     ██
 ██     ██ ██    ██ ██     ██
 ██  █  ██ ██    ██ ██  █  ██
 ██ ███ ██ ██    ██ ██ ███ ██  
  ███ ███   ██████   ███ ███   (c) 2025 Lewis Wynne

 Licensed under the GNU Affero General Public License v3

 Many flags support being written in shorthand, by using one dash
 and (usually) the first letter of the flag name. Shorthand flags
 can be combined by writing the letters together in any order. If
 a flag takes a value, it needs to be written last — you can only
 pass in one argument per command, so if you need to specify more,
 just write your flags separately.`

	ctx := kong.Must(&pda, kong.Description(description), kong.Bind(cfg))

	args := os.Args[1:]
	piped, _ := stdinHasData()
	k, err := ctx.Parse(args)

	if err != nil {
		defaultcmd := "get"

		if len(args) == 0 {
			defaultcmd = "list"

		}
		if piped {
			defaultcmd = "save"
		}
		args = append([]string{defaultcmd}, args...)
		k, err = ctx.Parse(args)
	}

	ctx.FatalIfErrorf(err)
	k.Run()
	ctx.FatalIfErrorf(err)

	return nil
}

func run() error {
	root()

	return nil
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

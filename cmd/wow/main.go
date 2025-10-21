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

	"github.com/alecthomas/kong"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

var Wow struct {
	Save   SaveCmd           `cmd:"1" aliases:"s" help:"Save a snippet."`
	Get    GetCmd            `cmd:"1" aliases:"g" help:"Get a snippet."`
	Edit   command.EditCmd   `cmd:"1" aliases:"e" help:"Edit a snippet."`
	Open   OpenCmd           `cmd:"1" aliases:"o" help:"Open a snippet."`
	List   ListCmd           `cmd:"1" aliases:"l,ls" help:"List snippets."`
	Remove command.RemoveCmd `cmd:"1" aliases:"r,rm" help:"Remove a snippet."`
}

type SaveCmd struct {
	Key string `arg:"" optional:"" name:"key" help:"Snippet key."`
	Tag string `help:"Comma-separated tags to add."`
}

func (c *SaveCmd) Run(ctx *kong.Context) error {
	return nil
}

type GetCmd struct {
	Key   string `arg:"" name:"key" help:"Snippet key."`
	Tag   string `help:"Comma-separated tags to add."`
	Untag string `help:"Comma-separated tags to remove."`
}

func (c *GetCmd) Run(ctx *kong.Context) error {
	return nil
}

type OpenCmd struct {
	Key   string `arg:"" name:"key" help:"Snippet key."`
	Pager bool   `help:"Prefer a CLI pager to xdg-open."`
}

func (c *OpenCmd) Run(ctx *kong.Context) error {
	return nil
}

type ListCmd struct {
	Plain   bool `help:"Format as a plain table of tab-separated values."`
	Tags    bool `short:"t" help:"Display tags."`
	Dates   bool `short:"D" help:"Display creation and last-modified dates."`
	Desc    bool `short:"d" help:"Display description."`
	Type    bool `short:"T" help:"Display type."`
	Verbose bool `short:"v" help:"Display all metadata."`
	Limit   int  `short:"l" default:"50" help:"Number of snippets per page."`
	Page    int  `short:"p" default:"1" help:"Page number."`
	All     bool `short:"a" help:"Disable pagination and display all snippets."`
}

func (c *ListCmd) Run(ctx *kong.Context, cfg command.Config) error {
	fmt.Print("sjakdasdjasdkadskj")
	return nil
}

func root() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := storage.InitMetaDB(cfg.MetaDB)
	if err != nil {
		return err
	}
	defer db.Close()

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

	description := `
 wow! a tool for code snippets

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

	ctx := kong.Must(&Wow, kong.Description(description), kong.Bind(cmdCfg))

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

	command.NewListCommand(cmdCfg).Execute([]string{})
	return nil
}

func run() error {
	root()

	return nil

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
	openCmd := command.NewOpenCommand(cmdCfg)
	listCmd := command.NewListCommand(cmdCfg)

	dispatcher.Register(saveCmd, "s")
	dispatcher.Register(getCmd, "g")
	dispatcher.Register(openCmd, "o")
	dispatcher.Register(listCmd, "ls")

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
  wow get    <key> [--tag str] [--untag str] [@tag] [-@tag]  Get a snippet.
  wow save   <key> [--tag str] [--desc str] [@tag]           Save a snippet.
  wow open   <key> [--pager]                                 Open a snippet. 
  wow edit   <key>                                           Edit a snippet.
  wow remove <key>                                           Remove a snippet.
  wow list [--limit int] [--page int] [--plain] [--verbose]  List snippets. 
           [--tags] [--type] [--desc] [--dates] [--all]
  wow help [command]                                         Get specific help.
  
  Run any command with --help for more info.

  Many flags support being written in shorthand, by using one dash
  and (usually) the first letter of the flag name. Shorthand flags
  can be combined by writing the letters together in any order. If
  a flag takes a value, it needs to be written last — you can only
  pass in one argument per command, so if you need to specify more,
  just write your flags separately.

  For example: "wow list -tdl 2" is --tags, --desc, and --limit 2.
`)
}

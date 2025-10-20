package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "wow",
		Short:              "A CLI tool for organising and tagging text snippets and bookmarks",
		SilenceErrors:      true,
		SilenceUsage:       true,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			piped, err := stdinHasData()
			if err != nil {
				return err
			}
			app, err := newApplication(cmd.InOrStdin(), cmd.OutOrStdout())
			if err != nil {
				return err
			}
			defer app.Close()
			return app.Execute(args, piped)
		},
	}

	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		printUsage(cmd.OutOrStdout())
		return nil
	})
	cmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		printUsage(cmd.OutOrStdout())
	})

	return cmd
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

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `Usage:
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

package main

import (
	"io"
	"time"

	"github.com/llywelwyn/wow/internal/command"
	"github.com/llywelwyn/wow/internal/config"
	"github.com/llywelwyn/wow/internal/editor"
	"github.com/llywelwyn/wow/internal/opener"
	"github.com/llywelwyn/wow/internal/pager"
	"github.com/llywelwyn/wow/internal/runner"
	"github.com/llywelwyn/wow/internal/storage"
)

type application struct {
	cfg        config.Config
	commandCfg command.Config

	dispatcher *command.Dispatcher

	saveCmd   *command.SaveCommand
	getCmd    *command.GetCommand
	editCmd   *command.EditCommand
	openCmd   *command.OpenCommand
	listCmd   *command.ListCommand
	removeCmd *command.RemoveCommand

	out   io.Writer
	close func() error
}

func newApplication(in io.Reader, out io.Writer) (*application, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	db, err := storage.InitMetaDB(cfg.MetaDB)
	if err != nil {
		return nil, err
	}

	cmdCfg := command.Config{
		BaseDir: cfg.BaseDir,
		DB:      db,
		Input:   in,
		Output:  out,
		Clock:   time.Now,
		Editor:  runner.Run(editor.GetEditorFromEnv()),
		Opener:  runner.Run(opener.GetOpenerFromEnv()),
		Pager:   runner.Run(pager.GetPagerFromEnv()),
	}

	app := &application{
		cfg:        cfg,
		commandCfg: cmdCfg,
		dispatcher: command.NewDispatcher(),
		saveCmd:    command.NewSaveCommand(cmdCfg),
		getCmd:     command.NewGetCommand(cmdCfg),
		editCmd:    command.NewEditCommand(cmdCfg),
		openCmd:    command.NewOpenCommand(cmdCfg),
		listCmd:    command.NewListCommand(cmdCfg),
		removeCmd:  command.NewRemoveCommand(cmdCfg),
		out:        out,
		close:      db.Close,
	}

	app.dispatcher.Register(app.saveCmd)
	app.dispatcher.Register(app.getCmd)
	app.dispatcher.Register(app.editCmd)
	app.dispatcher.Register(app.openCmd)
	app.dispatcher.Register(app.listCmd, "ls")
	app.dispatcher.Register(app.removeCmd, "rm")

	return app, nil
}

func (a *application) Close() error {
	if a == nil || a.close == nil {
		return nil
	}
	return a.close()
}

func (a *application) Execute(args []string, piped bool) error {
	if len(args) == 0 {
		if piped {
			return a.saveCmd.Execute(nil)
		}
		printUsage(a.out)
		return nil
	}

	if args[0] == "--help" || args[0] == "-h" {
		printUsage(a.out)
		return nil
	}

	if cmd, ok := a.dispatcher.Lookup(args[0]); ok {
		return cmd.Execute(args[1:])
	}

	if piped {
		return a.saveCmd.Execute(args)
	}
	return a.getCmd.Execute(args)
}

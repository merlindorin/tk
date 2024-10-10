package main

import (
	_ "embed"

	"github.com/alecthomas/kong"
	"github.com/merlindorin/tk/cmd/tk/commands"
	"github.com/merlindorin/tk/cmd/tk/global"
	"github.com/vanyda-official/go-shared/pkg/cmd"
)

//nolint:gochecknoglobals // these global variables exist to be overridden during build
var (
	name    = "tk"
	license string

	version     = "dev"
	commit      = "dirty"
	date        = "latest"
	buildSource = "source"
)

type CLI struct {
	*cmd.Commons
	*cmd.Config

	*global.Taskfiles

	Init commands.InitCmd `cmd:"init" help:"initialize a new workspace"`
}

func main() {
	cli := CLI{
		Commons: &cmd.Commons{
			Version: cmd.NewVersion(name, version, commit, buildSource, date),
			Licence: cmd.NewLicence(license),
		},
		Config:    cmd.NewConfig(name),
		Taskfiles: &global.Taskfiles{},
	}

	ctx := kong.Parse(
		&cli,
		kong.Name(name),
		kong.Description("Simple cli for managing my workspaces"),
		kong.UsageOnError(),
	)

	ctx.FatalIfErrorf(ctx.Run(cli.Taskfiles, cli.Commons))
}

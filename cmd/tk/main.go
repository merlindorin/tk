package main

import (
	_ "embed"

	"github.com/alecthomas/kong"
	"github.com/merlindorin/go-shared/pkg/cmd"

	"github.com/merlindorin/tk/cmd/tk/commands"
	"github.com/merlindorin/tk/cmd/tk/global"
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

	*global.TK

	Init   commands.InitCmd   `cmd:"init" help:"initialize a new workspace"`
	Update commands.UpdateCmd `cmd:"update" help:"update workspace"`
}

func main() {
	cli := CLI{
		Commons: &cmd.Commons{
			Version: cmd.NewVersion(name, version, commit, buildSource, date),
			Licence: cmd.NewLicence(license),
		},
		Config: cmd.NewConfig(name),
		TK:     &global.TK{},
	}

	ctx := kong.Parse(
		&cli,
		kong.Name(name),
		kong.Description("Simple cli for managing my workspaces"),
		kong.UsageOnError(),
	)

	ctx.FatalIfErrorf(ctx.Run(cli.TK, cli.Commons))
}

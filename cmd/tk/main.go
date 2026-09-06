package main

import (
	_ "embed"

	"github.com/alecthomas/kong"
	"github.com/merlindorin/go-shared/pkg/cmd"

	"github.com/merlindorin/tk/cmd/tk/commands"
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

	Init   commands.InitCmd   `cmd:"init" help:"initialize a new workspace"`
	Update commands.UpdateCmd `cmd:"update" help:"update workspace"`
	Status commands.StatusCmd `cmd:"status" help:"report how the workspace differs from the powerpacks"`
	List   commands.ListCmd   `cmd:"list" aliases:"ls" help:"list the available powerpacks"`
	Add    commands.AddCmd    `cmd:"add" help:"install powerpacks in the workspace"`
	Remove commands.RemoveCmd `cmd:"remove" aliases:"rm" help:"remove powerpacks from the workspace"`
}

func main() {
	cli := CLI{
		Commons: &cmd.Commons{
			Version: cmd.NewVersion(name, version, commit, buildSource, date),
			Licence: cmd.NewLicence(license),
		},
		Config: cmd.NewConfig(name),
	}

	ctx := kong.Parse(
		&cli,
		kong.Name(name),
		kong.Description("Simple cli for managing my workspaces"),
		kong.UsageOnError(),
	)

	ctx.FatalIfErrorf(ctx.Run(cli.Commons))
}

package commands

import (
	"fmt"

	"github.com/merlindorin/go-shared/pkg/cmd"
	ps "github.com/merlindorin/tk/pkg/powerpacks"
	"github.com/merlindorin/tk/powerpacks"
)

type InitCmd struct {
	Target string `help:"target where to init tk" default:"."`

	Exclude []string `help:"exclude powerpacks" optional:""`

	DisableEnvrc     bool `default:"false" help:"disable envrc"`
	DisableTaskfiles bool `default:"false" help:"disable taskfiles"`
	DisableAqua      bool `default:"false" help:"disable aqua"`
	DisableReadme    bool
}

func (i *InitCmd) Run(_ *cmd.Commons) error {
	manager, err := powerpacks.BuildPowerpackManager()
	if err != nil {
		return fmt.Errorf("failed to create powerpacks: %w", err)
	}

	opts := ps.WriteOption{
		IgnoreAqua:     i.DisableAqua,
		IgnoreTaskfile: i.DisableTaskfiles,
		IgnoreReadme:   i.DisableReadme,
		Excludes:       i.Exclude,
	}

	if er := manager.Write(".", opts); er != nil {
		return fmt.Errorf("failed to write powerpacks: %w", er)
	}

	return nil
}

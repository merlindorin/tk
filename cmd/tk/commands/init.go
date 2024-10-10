package commands

import (
	"fmt"
	"time"

	ps "github.com/merlindorin/tk/pkg/powerpacks"

	"github.com/merlindorin/tk/powerpacks"

	"github.com/vanyda-official/go-shared/pkg/cmd"
)

type InitCmd struct {
	Name        string `arg:"" optional:"" help:"project name"`
	Description string `help:"project description"`

	DisableEnvrc     bool `default:"false" help:"disable envrc"`
	DisableTaskfiles bool `default:"false" help:"disable taskfiles"`
	DisableAqua      bool `default:"false" help:"disable aqua"`
	DisableReadme    bool

	Force bool `default:"false" help:"overwrite existing files if already exist"`

	Timeout time.Duration `default:"5s" help:"timeout for discovering (ns, ms, s & m)"`
}

func (i *InitCmd) Run(_ *cmd.Commons) error {
	p, err := powerpacks.BuildPowerpackManager()
	if err != nil {
		return fmt.Errorf("failed to create powerpacks: %w", err)
	}

	opts := ps.WriteOption{
		IgnoreAqua:     i.DisableAqua,
		IgnoreTaskfile: i.DisableTaskfiles,
		IgnoreReadme:   i.DisableReadme,
	}

	if er := p.Write(".", opts); er != nil {
		return fmt.Errorf("failed to write powerpacks: %w", er)
	}

	return nil
}

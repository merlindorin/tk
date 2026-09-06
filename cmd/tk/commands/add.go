package commands

import (
	"fmt"

	"github.com/merlindorin/go-shared/pkg/cmd"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

// AddCmd opts powerpacks back into a project.
type AddCmd struct {
	mutation

	Powerpacks []string `arg:"" help:"powerpacks to install" required:""`
}

// Run installs the powerpacks and reports the changes.
func (a *AddCmd) Run(commons *cmd.Commons) error {
	manager, err := a.manager()
	if err != nil {
		return err
	}

	if err = manager.Validate(a.Powerpacks...); err != nil {
		return err
	}

	config, err := ps.LoadConfig(a.Target)
	if err != nil {
		return err
	}

	config.Version = version(commons)

	added := manager.Enable(&config, a.Powerpacks...)
	if len(added) == 0 {
		if _, err = fmt.Fprintf(a.writer(), "already installed\n"); err != nil {
			return fmt.Errorf("failed to report changes: %w", err)
		}
	}

	if len(added) > 0 {
		if _, err = fmt.Fprintf(a.writer(), "installing %v\n", added); err != nil {
			return fmt.Errorf("failed to report changes: %w", err)
		}
	}

	return a.run(manager, config)
}

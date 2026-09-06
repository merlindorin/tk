package commands

import (
	"fmt"

	"github.com/merlindorin/go-shared/pkg/cmd"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

// RemoveCmd opts powerpacks out of a project.
type RemoveCmd struct {
	mutation

	Powerpacks []string `arg:"" help:"powerpacks to remove" required:""`
}

// Run removes the powerpacks and reports the changes.
func (r *RemoveCmd) Run(commons *cmd.Commons) error {
	manager, err := r.manager()
	if err != nil {
		return err
	}

	if err = manager.Validate(r.Powerpacks...); err != nil {
		return err
	}

	config, err := ps.LoadConfig(r.Target)
	if err != nil {
		return err
	}

	config.Version = version(commons)

	removed, err := manager.Disable(&config, r.Powerpacks...)
	if err != nil {
		return err
	}

	if len(removed) > 0 {
		if _, err = fmt.Fprintf(r.writer(), "removing %v\n", removed); err != nil {
			return fmt.Errorf("failed to report changes: %w", err)
		}
	}

	return r.run(manager, config)
}

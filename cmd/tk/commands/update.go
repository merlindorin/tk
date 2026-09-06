package commands

import (
	"github.com/merlindorin/go-shared/pkg/cmd"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

// UpdateCmd refreshes a project from the powerpacks embedded in the binary.
type UpdateCmd struct {
	mutation
}

// Run updates the project and reports the changes.
func (u *UpdateCmd) Run(commons *cmd.Commons) error {
	manager, err := u.manager()
	if err != nil {
		return err
	}

	config, err := ps.LoadConfig(u.Target)
	if err != nil {
		return err
	}

	config.Version = version(commons)

	return u.run(manager, config)
}

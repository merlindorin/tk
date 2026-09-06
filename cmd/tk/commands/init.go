package commands

import (
	"errors"
	"fmt"

	"github.com/merlindorin/go-shared/pkg/cmd"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

// InitCmd installs tk in a project for the first time.
type InitCmd struct {
	mutation

	Include []string `help:"powerpacks to install; every one of them when left out" optional:""`

	DisableEnvrc     bool `help:"leave .envrc alone, including the export tk used to write there"`
	DisableTaskfiles bool `help:"do not manage Taskfiles"`
	DisableReadme    bool `help:"do not write the powerpack documentation"`
}

// Run initialises the project and reports the changes.
func (i *InitCmd) Run(commons *cmd.Commons) error {
	manager, err := i.manager()
	if err != nil {
		return err
	}

	if err = manager.Validate(i.Include...); err != nil {
		return err
	}

	if err = i.warnOnReinit(); err != nil {
		return err
	}

	config := ps.Config{
		Version:        version(commons),
		IgnoreReadme:   i.DisableReadme,
		IgnoreTaskfile: i.DisableTaskfiles,
		IgnoreEnvrc:    i.DisableEnvrc,
		Includes:       i.Include,
		Excludes:       nil,
		Powerpacks:     nil,
	}

	return i.run(manager, config)
}

// warnOnReinit tells the user when an existing configuration is about to be replaced,
// since `tk init` rebuilds it from the flags rather than from what the project had.
func (i *InitCmd) warnOnReinit() error {
	config, err := ps.LoadConfig(i.Target)
	if errors.Is(err, ps.ErrConfigNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	if _, err = fmt.Fprintf(i.writer(),
		"%s already exists and is being replaced; use `tk update`, `tk add` or `tk remove` to keep it\n",
		ps.ConfigFilename); err != nil {
		return fmt.Errorf("failed to report changes: %w", err)
	}

	if !config.SelectsAll() && len(i.Include) == 0 {
		if _, err = fmt.Fprintf(i.writer(), "installing every powerpack, it selected only: %v\n",
			selection(&config)); err != nil {
			return fmt.Errorf("failed to report changes: %w", err)
		}
	}

	return nil
}

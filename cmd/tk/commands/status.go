package commands

import (
	"errors"
	"fmt"

	"github.com/merlindorin/go-shared/pkg/cmd"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

// ErrOutOfDate is reported by `tk status --check` when the project needs an update.
var ErrOutOfDate = errors.New("project is out of date")

// StatusCmd reports how a project differs from the powerpacks embedded in the binary.
type StatusCmd struct {
	project

	Check bool `help:"exit with a non-zero status when the project is out of date"`
}

// Run reports the state of the project without changing anything.
func (s *StatusCmd) Run(commons *cmd.Commons) error {
	manager, err := s.manager()
	if err != nil {
		return err
	}

	config, err := ps.LoadConfig(s.Target)
	if err != nil {
		return err
	}

	if err = s.describe(&config, len(manager.Selected(config)), version(commons)); err != nil {
		return err
	}

	plan, err := manager.Plan(s.Target, config)
	if err != nil {
		return err
	}

	if err = report(s.writer(), plan, true); err != nil {
		return err
	}

	if plan.IsUpToDate() {
		return nil
	}

	if _, err = fmt.Fprintf(s.writer(), "run `tk update` to apply\n"); err != nil {
		return fmt.Errorf("failed to report status: %w", err)
	}

	if s.Check {
		return fmt.Errorf("%w: %d changes pending", ErrOutOfDate, len(plan.Pending()))
	}

	return nil
}

// describe prints where the project stands: which tk wrote it, and what it installs.
func (s *StatusCmd) describe(config *ps.Config, selected int, running string) error {
	installedWith := config.Version
	if installedWith == "" {
		installedWith = "unknown"
	}

	if _, err := fmt.Fprintf(s.writer(), "tk %s (project written with %s)\n", running, installedWith); err != nil {
		return fmt.Errorf("failed to report status: %w", err)
	}

	scope := "every powerpack tk ships"
	if !config.SelectsAll() {
		scope = fmt.Sprintf("selected: %v", selection(config))
	}

	_, err := fmt.Fprintf(s.writer(), "%s installed, %s\n\n",
		plural(selected, "powerpack", "powerpacks"), scope)
	if err != nil {
		return fmt.Errorf("failed to report status: %w", err)
	}

	return nil
}

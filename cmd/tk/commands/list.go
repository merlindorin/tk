package commands

import (
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

// ListCmd shows the powerpacks shipped with the binary and their state in the project.
type ListCmd struct {
	project

	Tasks bool `help:"list the tasks each powerpack defines instead of its description"`
}

// Run lists the powerpacks.
func (l *ListCmd) Run() error {
	manager, err := l.manager()
	if err != nil {
		return err
	}

	config, err := ps.LoadConfig(l.Target)
	if err != nil && !errors.Is(err, ps.ErrConfigNotFound) {
		return err
	}

	installed := !errors.Is(err, ps.ErrConfigNotFound)

	table := tabwriter.NewWriter(l.writer(), 0, 0, 2, ' ', 0) //nolint:mnd // padding of the powerpack list

	if err = l.header(table); err != nil {
		return err
	}

	for _, powerpack := range manager.List() {
		if _, err = fmt.Fprintf(table, "%s\t%s\t%s\t%s\n",
			powerpack.Name, powerpack.Prefix(), state(&config, &powerpack, installed),
			l.details(&powerpack)); err != nil {
			return fmt.Errorf("failed to list powerpacks: %w", err)
		}
	}

	if err = table.Flush(); err != nil {
		return fmt.Errorf("failed to list powerpacks: %w", err)
	}

	return nil
}

// header names the columns, the prefix one being how the tasks are actually invoked.
func (l *ListCmd) header(table *tabwriter.Writer) error {
	last := "DESCRIPTION"
	if l.Tasks {
		last = "TASKS"
	}

	if _, err := fmt.Fprintf(table, "NAME\tPREFIX\tSTATE\t%s\n", last); err != nil {
		return fmt.Errorf("failed to list powerpacks: %w", err)
	}

	return nil
}

// details is the last column: what the powerpack is, or what it lets you run.
func (l *ListCmd) details(powerpack *ps.Powerpack) string {
	if !l.Tasks {
		return powerpack.Summary()
	}

	tasks := powerpack.Tasks()
	for i, task := range tasks {
		tasks[i] = powerpack.Prefix() + task
	}

	return strings.Join(tasks, " ")
}

// state describes what a powerpack is doing in this project.
func state(config *ps.Config, powerpack *ps.Powerpack, installed bool) string {
	switch {
	case !installed, !config.IsSelected(powerpack.Name):
		return "available"
	case len(config.Powerpacks) == 0:
		// A configuration written before tk recorded provenance.
		return "installed"
	case config.Powerpacks[powerpack.Name] == "":
		// Shipped by a newer tk than the one that wrote the project.
		return "new"
	case config.Powerpacks[powerpack.Name] != powerpack.Checksum():
		return "outdated"
	default:
		return "installed"
	}
}

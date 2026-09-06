// Package commands implements the tk sub-commands.
package commands

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/merlindorin/go-shared/pkg/cmd"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
	"github.com/merlindorin/tk/powerpacks"
)

// project is the set of flags every command shares.
type project struct {
	Target string `help:"project directory to work in" default:"." type:"path"`

	Out io.Writer `kong:"-"`
}

// writer is where the command reports what it did.
func (p *project) writer() io.Writer {
	if p.Out == nil {
		return os.Stdout
	}

	return p.Out
}

// manager loads the powerpacks embedded in the binary.
func (p *project) manager() (*ps.Manager, error) {
	manager, err := powerpacks.BuildPowerpackManager()
	if err != nil {
		return nil, fmt.Errorf("failed to load powerpacks: %w", err)
	}

	return manager, nil
}

// mutation is the set of flags shared by the commands that write to the project.
type mutation struct {
	project

	DryRun bool `name:"dry-run" help:"report the changes without writing anything"`
}

// run applies a plan unless the command is a dry run, then reports what happened.
func (m *mutation) run(manager *ps.Manager, config ps.Config) error {
	plan, err := manager.Plan(m.Target, config)
	if err != nil {
		return err
	}

	if !m.DryRun {
		if err = plan.Apply(); err != nil {
			return err
		}
	}

	return report(m.writer(), plan, m.DryRun)
}

// version returns the tk version the project is being written with.
func version(commons *cmd.Commons) string {
	if commons == nil {
		return ""
	}

	return commons.Version.Version()
}

// report prints the changes of a plan.
func report(writer io.Writer, plan *ps.Plan, dryRun bool) error {
	pending := plan.Pending()
	unchanged := len(plan.Changes) - len(pending)

	table := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0) //nolint:mnd // padding of the change list

	for _, change := range pending {
		if _, err := fmt.Fprintf(table, "  %s\t%s\n", change.Action, change.Path); err != nil {
			return fmt.Errorf("failed to report changes: %w", err)
		}
	}

	if err := table.Flush(); err != nil {
		return fmt.Errorf("failed to report changes: %w", err)
	}

	if err := notes(writer, plan.Notes); err != nil {
		return err
	}

	return summarize(writer, len(pending), unchanged, dryRun)
}

// notes prints what tk noticed but will not do on its own, so a file a powerpack left
// behind is visible rather than silently deleted or silently kept.
func notes(writer io.Writer, lines []string) error {
	for _, line := range lines {
		if _, err := fmt.Fprintf(writer, "  note: %s\n", line); err != nil {
			return fmt.Errorf("failed to report changes: %w", err)
		}
	}

	return nil
}

// summarize prints the last line of a report.
func summarize(writer io.Writer, pending, unchanged int, dryRun bool) error {
	var err error

	switch {
	case pending == 0:
		_, err = fmt.Fprintf(writer, "already up to date (%d files)\n", unchanged)
	case dryRun:
		_, err = fmt.Fprintf(writer, "\n%s to apply, %d unchanged (dry run, nothing written)\n",
			plural(pending, "change", "changes"), unchanged)
	default:
		_, err = fmt.Fprintf(writer, "\n%s applied, %d unchanged\n",
			plural(pending, "change", "changes"), unchanged)
	}

	if err != nil {
		return fmt.Errorf("failed to report changes: %w", err)
	}

	return nil
}

// selection renders the powerpacks a project asked for, whatever list it uses.
func selection(config *ps.Config) []string {
	if len(config.Includes) > 0 {
		return config.Includes
	}

	return config.Excludes
}

// plural renders a count with the wording matching it.
func plural(count int, singular, many string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}

	return fmt.Sprintf("%d %s", count, many)
}

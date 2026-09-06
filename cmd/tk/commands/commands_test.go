package commands_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/merlindorin/go-shared/pkg/cmd"
	"gopkg.in/yaml.v3"

	"github.com/merlindorin/tk/cmd/tk/commands"
	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

func commons() *cmd.Commons {
	return &cmd.Commons{
		Development: false, Level: "info", Lang: "en",
		Version: cmd.NewVersion("tk", "v1.2.3", "commit", "test", "date"),
		Licence: cmd.NewLicence(""),
	}
}

func initProject(t *testing.T, target string, include ...string) string {
	t.Helper()

	out := &bytes.Buffer{}
	command := &commands.InitCmd{
		Include:          include,
		DisableTaskfiles: false, DisableReadme: false,
	}
	command.Target = target
	command.Out = out

	assert.NoError(t, command.Run(commons()))

	return out.String()
}

func TestInitCmd(t *testing.T) {
	target := t.TempDir()
	out := initProject(t, target)

	assert.Contains(t, out, "create")
	assert.Contains(t, out, "changes applied")

	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)
	assert.Equal(t, "v1.2.3", config.Version)

	_, err = os.Stat(filepath.Join(target, "Taskfile.yaml"))
	assert.NoError(t, err)
}

func TestInitCmd_dryRunWritesNothing(t *testing.T) {
	target := t.TempDir()

	out := &bytes.Buffer{}
	command := &commands.InitCmd{
		Include: nil, DisableTaskfiles: false, DisableReadme: false,
	}
	command.Target = target
	command.Out = out
	command.DryRun = true

	assert.NoError(t, command.Run(commons()))
	assert.Contains(t, out.String(), "dry run")

	entries, err := os.ReadDir(target)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(entries))
}

func TestInitCmd_rejectsUnknownPowerpack(t *testing.T) {
	command := &commands.InitCmd{
		Include: []string{"nope"}, DisableTaskfiles: false, DisableReadme: false,
	}
	command.Target = t.TempDir()
	command.Out = &bytes.Buffer{}

	assert.IsError(t, command.Run(commons()), ps.ErrUnknownPowerpack)
}

func TestUpdateCmd_isIdempotent(t *testing.T) {
	target := t.TempDir()
	initProject(t, target)

	out := &bytes.Buffer{}
	command := &commands.UpdateCmd{}
	command.Target = target
	command.Out = out

	assert.NoError(t, command.Run(commons()))
	assert.Contains(t, out.String(), "already up to date")
}

func TestUpdateCmd_withoutConfig(t *testing.T) {
	command := &commands.UpdateCmd{}
	command.Target = t.TempDir()
	command.Out = &bytes.Buffer{}

	assert.IsError(t, command.Run(commons()), ps.ErrConfigNotFound)
}

func TestAddAndRemoveCmd(t *testing.T) {
	target := t.TempDir()
	initProject(t, target)

	out := &bytes.Buffer{}
	remove := &commands.RemoveCmd{Powerpacks: []string{"golangci"}}
	remove.Target = target
	remove.Out = out

	assert.NoError(t, remove.Run(commons()))
	assert.Contains(t, out.String(), "delete")

	_, err := os.Stat(filepath.Join(target, ".tk", "golangci"))
	assert.Error(t, err)

	out.Reset()

	add := &commands.AddCmd{Powerpacks: []string{"golangci"}}
	add.Target = target
	add.Out = out

	assert.NoError(t, add.Run(commons()))
	assert.Contains(t, out.String(), "create")

	_, err = os.Stat(filepath.Join(target, ".tk", "golangci", "Taskfile.yaml"))
	assert.NoError(t, err)
}

func TestRemoveCmd_rejectsUnknownPowerpack(t *testing.T) {
	target := t.TempDir()
	initProject(t, target)

	command := &commands.RemoveCmd{Powerpacks: []string{"nope"}}
	command.Target = target
	command.Out = &bytes.Buffer{}

	assert.IsError(t, command.Run(commons()), ps.ErrUnknownPowerpack)
}

func TestListCmd(t *testing.T) {
	target := t.TempDir()
	initProject(t, target, "git", "golangci")

	out := &bytes.Buffer{}
	command := &commands.ListCmd{Tasks: false}
	command.Target = target
	command.Out = out

	assert.NoError(t, command.Run())
	assert.Contains(t, out.String(), "NAME")
	assert.Contains(t, out.String(), "PREFIX")
	assert.Contains(t, out.String(), "git         git:")
	assert.Contains(t, out.String(), "installed")
	assert.Contains(t, out.String(), "available", "a powerpack left out of the selection")
}

func TestListCmd_withTasks(t *testing.T) {
	target := t.TempDir()
	initProject(t, target)

	out := &bytes.Buffer{}
	command := &commands.ListCmd{Tasks: true}
	command.Target = target
	command.Out = out

	assert.NoError(t, command.Run())
	assert.Contains(t, out.String(), "TASKS")
	assert.Contains(t, out.String(), "golangci:lint")
	assert.Contains(t, out.String(), "git:install")
}

func TestListCmd_withoutConfig(t *testing.T) {
	out := &bytes.Buffer{}
	command := &commands.ListCmd{Tasks: false}
	command.Target = t.TempDir()
	command.Out = out

	assert.NoError(t, command.Run())
	assert.Contains(t, out.String(), "available")
}

func TestInitCmd_withASelection(t *testing.T) {
	target := t.TempDir()
	initProject(t, target, "git")

	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)
	assert.Equal(t, []string{"git"}, config.Includes)
	assert.True(t, config.IsSelected("git"))
	assert.False(t, config.IsSelected("golangci"))

	_, err = os.Stat(filepath.Join(target, ".tk", "golangci"))
	assert.Error(t, err, "only the selected powerpack is written")

	assert.Contains(t, read(t, filepath.Join(target, "Taskfile.yaml")), "git: .tk/git/Taskfile.yaml #!tk")
}

func read(t *testing.T, name string) string {
	t.Helper()

	content, err := os.ReadFile(name)
	assert.NoError(t, err)

	return string(content)
}

func TestStatusCmd(t *testing.T) {
	target := t.TempDir()
	initProject(t, target)

	out := &bytes.Buffer{}
	command := &commands.StatusCmd{}
	command.Target = target
	command.Out = out

	assert.NoError(t, command.Run(commons()))
	assert.Contains(t, out.String(), "tk v1.2.3")
	assert.Contains(t, out.String(), "already up to date")
}

func TestStatusCmd_reportsDrift(t *testing.T) {
	target := t.TempDir()
	initProject(t, target)

	// A hand-edited generated file must show up as drift.
	generated := filepath.Join(target, ".tk", "git", "Taskfile.yaml")
	assert.NoError(t, os.WriteFile(generated, []byte("version: '3'\n"), 0o600))

	out := &bytes.Buffer{}
	command := &commands.StatusCmd{}
	command.Target = target
	command.Out = out

	assert.NoError(t, command.Run(commons()))
	assert.Contains(t, out.String(), "update")
	assert.Contains(t, out.String(), "run `tk update` to apply")
}

func TestInitCmd_warnsWhenReinitialising(t *testing.T) {
	target := t.TempDir()
	initProject(t, target, "claude")

	out := initProject(t, target)
	assert.Contains(t, out, "already exists and is being replaced")
	assert.Contains(t, out, "installing every powerpack, it selected only: [claude]")

	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)
	assert.True(t, config.IsSelected("golangci"))
}

func TestStatusCmd_checkFailsOnDrift(t *testing.T) {
	target := t.TempDir()
	initProject(t, target)

	command := &commands.StatusCmd{Check: true}
	command.Target = target
	command.Out = &bytes.Buffer{}

	assert.NoError(t, command.Run(commons()), "an up to date project passes the check")

	assert.NoError(t, os.WriteFile(filepath.Join(target, ".tk", "git", "Taskfile.yaml"), []byte("drift\n"), 0o600))
	assert.IsError(t, command.Run(commons()), commands.ErrOutOfDate)
}

func TestListCmd_reportsNewAndOutdatedPowerpacks(t *testing.T) {
	target := t.TempDir()
	initProject(t, target)

	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)

	// A powerpack the project has never seen, and one whose content moved on.
	delete(config.Powerpacks, "git")
	config.Powerpacks["golangci"] = "sha256:stale"
	assert.NoError(t, os.WriteFile(filepath.Join(target, ".tk.yaml"), marshal(t, config), 0o600))

	out := &bytes.Buffer{}
	command := &commands.ListCmd{Tasks: false}
	command.Target = target
	command.Out = out

	assert.NoError(t, command.Run())
	assert.Contains(t, out.String(), "new")
	assert.Contains(t, out.String(), "outdated")
}

func marshal(t *testing.T, config ps.Config) []byte {
	t.Helper()

	content, err := yaml.Marshal(config)
	assert.NoError(t, err)

	return content
}

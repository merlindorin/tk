package powerpacks_test

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

// owningManager returns a manager with a powerpack that writes a workflow itself and one
// whose tasks write a configuration file, which is the whole point of the two strategies.
func owningManager(t *testing.T) *ps.Manager {
	t.Helper()

	synced, err := ps.ParseManifest("ci", []byte(`
owns:
  - path: .github/workflows/ci.yml
    strategy: sync
    source: files/workflow.yml
`))
	assert.NoError(t, err)

	generated, err := ps.ParseManifest("linter", []byte(`
owns:
  - path: .linter.yaml
`))
	assert.NoError(t, err)

	manager := ps.NewPowerpackManager()
	manager.Add(&ps.Powerpack{
		Name: "ci", Description: "",
		Taskfile: []byte("version: '3'\ntasks:\n  run: {}\n"), Readme: []byte("# CI\n"),
		Manifest: synced, Sources: map[string][]byte{"files/workflow.yml": []byte("name: ci\n")},
	})
	manager.Add(&ps.Powerpack{
		Name: "linter", Description: "",
		Taskfile: []byte("version: '3'\ntasks:\n  lint: {}\n"), Readme: []byte("# Linter\n"),
		Manifest: generated, Sources: nil,
	})

	return manager
}

func TestOwnership_syncWritesAndUpdates(t *testing.T) {
	target := t.TempDir()
	manager := owningManager(t)

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)
	assert.Equal(t, "name: ci\n", read(t, target, ".github", "workflows", "ci.yml"))

	// The project edits it; tk owns the content, so an update puts it back.
	write(t, target, ".github/workflows/ci.yml", "name: drifted\n")

	_, err = manager.Write(target, testConfig())
	assert.NoError(t, err)
	assert.Equal(t, "name: ci\n", read(t, target, ".github", "workflows", "ci.yml"))
}

func TestOwnership_syncIsRecorded(t *testing.T) {
	target := t.TempDir()

	_, err := owningManager(t).Write(target, testConfig())
	assert.NoError(t, err)

	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)
	assert.Contains(t, config.Generated[".github/workflows/ci.yml"], "sha256:")
	assert.Zero(t, config.Generated[".linter.yaml"], "tk did not write that one")
}

func TestOwnership_syncIsRemovedWithItsPowerpack(t *testing.T) {
	target := t.TempDir()
	manager := owningManager(t)

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)

	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)

	_, err = manager.Disable(&config, "ci")
	assert.NoError(t, err)

	plan, err := manager.Write(target, config)
	assert.NoError(t, err)
	assert.False(t, exists(target, ".github", "workflows", "ci.yml"))
	assert.Zero(t, len(plan.Notes), "tk wrote it and it was untouched, so it just goes")
}

func TestOwnership_anEditedSyncFileIsKeptAndReported(t *testing.T) {
	target := t.TempDir()
	manager := owningManager(t)

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)

	write(t, target, ".github/workflows/ci.yml", "name: mine now\n")

	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)

	_, err = manager.Disable(&config, "ci")
	assert.NoError(t, err)

	plan, err := manager.Write(target, config)
	assert.NoError(t, err)
	assert.Equal(t, "name: mine now\n", read(t, target, ".github", "workflows", "ci.yml"))
	assert.Equal(t, 1, len(plan.Notes))
	assert.Contains(t, plan.Notes[0], "edited since tk wrote it")
}

func TestOwnership_onceIsNeverDeleted(t *testing.T) {
	// tk never wrote the content, so it cannot know that removing it is what is wanted.
	target := t.TempDir()
	manager := owningManager(t)

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)

	write(t, target, ".linter.yaml", "rules: []\n")

	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)

	_, err = manager.Disable(&config, "linter")
	assert.NoError(t, err)

	plan, err := manager.Write(target, config)
	assert.NoError(t, err)
	assert.True(t, exists(target, ".linter.yaml"))
	assert.Equal(t, 1, len(plan.Notes))
	assert.Contains(t, plan.Notes[0], "written by its tasks")
}

func TestOwnership_reportsAFileNoPowerpackClaimsAnymore(t *testing.T) {
	target := t.TempDir()
	manager := owningManager(t)

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)

	// A powerpack that stopped shipping a file, or was dropped from tk entirely.
	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)

	config.Generated["forgotten.yml"] = "sha256:whatever"
	write(t, target, "forgotten.yml", "left over\n")

	plan, err := manager.Plan(target, config)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(plan.Notes))
	assert.Contains(t, plan.Notes[0], "no longer claimed by any powerpack")
}

func TestRequires_pullsInWhatAPowerpackNeeds(t *testing.T) {
	manifest, err := ps.ParseManifest("ci", []byte("requires: [linter]\n"))
	assert.NoError(t, err)

	manager := owningManager(t)
	manager.Add(&ps.Powerpack{
		Name: "ci", Description: "",
		Taskfile: []byte("version: '3'\ntasks:\n  run: {}\n"), Readme: []byte("# CI\n"),
		Manifest: manifest, Sources: nil,
	})

	config := testConfig()
	config.Includes = []string{"ci"}

	selected := strings.Join(namesOf(manager.Selected(config)), " ")
	assert.Equal(t, "ci linter", selected)
	assert.True(t, manager.IsInstalled(config, "linter"), "pulled in, so installed")
	assert.False(t, config.IsSelected("linter"), "but not what the project named")

	assert.Equal(t, []string{"ci"}, manager.RequiredBy(config, "linter"))

	_, err = manager.Disable(&config, "linter")
	assert.IsError(t, err, ps.ErrRequiredPowerpack)
}

func TestRequires_removingTheDependentFreesTheRequirement(t *testing.T) {
	manifest, err := ps.ParseManifest("ci", []byte("requires: [linter]\n"))
	assert.NoError(t, err)

	manager := owningManager(t)
	manager.Add(&ps.Powerpack{
		Name: "ci", Description: "",
		Taskfile: []byte("version: '3'\ntasks:\n  run: {}\n"), Readme: []byte("# CI\n"),
		Manifest: manifest, Sources: nil,
	})

	config := testConfig()
	config.Includes = []string{"ci"}

	removed, err := manager.Disable(&config, "ci")
	assert.NoError(t, err)
	assert.Equal(t, []string{"ci"}, removed)
	assert.Equal(t, []string{"linter"}, config.Includes, "what it required stays, written out")
}

func namesOf(list []ps.Powerpack) []string {
	names := make([]string, 0, len(list))
	for i := range list {
		names = append(names, list[i].Name)
	}

	return names
}

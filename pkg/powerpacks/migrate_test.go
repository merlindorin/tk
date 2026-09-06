package powerpacks_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/alecthomas/assert/v2"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

// migratingManager returns a manager whose powerpack declares the migration under test.
func migratingManager(t *testing.T, manifest string) *ps.Manager {
	t.Helper()

	parsed, err := ps.ParseManifest("git", []byte(manifest))
	assert.NoError(t, err)

	manager := ps.NewPowerpackManager()
	manager.Add(&ps.Powerpack{
		Name: "git", Description: "",
		Taskfile: []byte("version: '3'\ntasks:\n  hook: {}\n"), Readme: []byte("# Git\n"),
		Manifest: parsed, Sources: nil,
	})

	return manager
}

func write(t *testing.T, target, name, content string) {
	t.Helper()

	assert.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(target, name)), 0o750))
	assert.NoError(t, os.WriteFile(filepath.Join(target, name), []byte(content), 0o600))
}

func TestMigration_removeLines(t *testing.T) {
	target := t.TempDir()
	write(t, target, "legacy.sh", "keep me\nexport OLD=1\nkeep me too\n")

	manager := migratingManager(t, `
migrations:
  - id: drop-old
    reason: it is not needed anymore
    remove-lines:
      path: legacy.sh
      matching: '^export OLD='
`)

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)
	assert.Equal(t, "keep me\nkeep me too\n", read(t, target, "legacy.sh"))

	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)
	assert.True(t, slices.Contains(config.Migrations, "git:drop-old"))
}

func TestMigration_removeLinesDeletesAnEmptiedFile(t *testing.T) {
	target := t.TempDir()
	write(t, target, "legacy.sh", "export OLD=1\n")

	manager := migratingManager(t, `
migrations:
  - id: drop-old
    remove-lines:
      path: legacy.sh
      matching: '^export OLD='
      delete-when-empty: true
`)

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)
	assert.False(t, exists(target, "legacy.sh"))
}

func TestMigration_removeFile(t *testing.T) {
	target := t.TempDir()
	write(t, target, ".github/workflows/old.yml", "name: old\n")

	manager := migratingManager(t, `
migrations:
  - id: drop-workflow
    remove-file: .github/workflows/old.yml
`)

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)
	assert.False(t, exists(target, ".github", "workflows", "old.yml"))
}

func TestMigration_renameFile(t *testing.T) {
	target := t.TempDir()
	write(t, target, ".github/workflows/old.yml", "name: kept\n")

	manager := migratingManager(t, `
migrations:
  - id: rename-workflow
    rename-file:
      from: .github/workflows/old.yml
      to: .github/workflows/new.yml
`)

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)
	assert.False(t, exists(target, ".github", "workflows", "old.yml"))
	assert.Equal(t, "name: kept\n", read(t, target, ".github", "workflows", "new.yml"))
}

func TestMigration_renameKeepsAnExistingDestination(t *testing.T) {
	// The project already moved on, by hand or through an earlier tk: do not clobber it.
	target := t.TempDir()
	write(t, target, ".github/workflows/old.yml", "name: old\n")
	write(t, target, ".github/workflows/new.yml", "name: mine\n")

	manager := migratingManager(t, `
migrations:
  - id: rename-workflow
    rename-file:
      from: .github/workflows/old.yml
      to: .github/workflows/new.yml
`)

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)
	assert.Equal(t, "name: mine\n", read(t, target, ".github", "workflows", "new.yml"))
	assert.True(t, exists(target, ".github", "workflows", "old.yml"))
}

func TestMigration_runsOnce(t *testing.T) {
	target := t.TempDir()
	write(t, target, "legacy.sh", "export OLD=1\n")

	manifest := `
migrations:
  - id: drop-old
    remove-lines:
      path: legacy.sh
      matching: '^export OLD='
`
	manager := migratingManager(t, manifest)

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)

	// The project puts the line back: a migration that already ran must leave it alone.
	write(t, target, "legacy.sh", "export OLD=1\n")

	// As `tk update` does, the second run starts from the configuration on disk.
	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)

	plan, err := manager.Write(target, config)
	assert.NoError(t, err)
	assert.True(t, plan.IsUpToDate())
	assert.Equal(t, "export OLD=1\n", read(t, target, "legacy.sh"))
}

func TestMigration_isPlannedNotApplied(t *testing.T) {
	target := t.TempDir()
	write(t, target, "legacy.sh", "export OLD=1\n")

	manager := migratingManager(t, `
migrations:
  - id: drop-old
    remove-lines:
      path: legacy.sh
      matching: '^export OLD='
      delete-when-empty: true
`)

	plan, err := manager.Plan(target, testConfig())
	assert.NoError(t, err)
	assert.True(t, exists(target, "legacy.sh"), "planning writes nothing")

	deletes := 0

	for _, change := range plan.Pending() {
		if change.Path == "legacy.sh" && change.Action == ps.ActionDelete {
			deletes++
		}
	}

	assert.Equal(t, 1, deletes, "the migration shows up in the plan, once")
}

func TestCoreMigration_dropsTheRemoteTaskfilesExport(t *testing.T) {
	target := t.TempDir()
	write(t, target, ".envrc", "use flake\nexport TASK_X_REMOTE_TASKFILES=1\n")

	_, err := testManager().Write(target, testConfig())
	assert.NoError(t, err)
	assert.Equal(t, "use flake\n", read(t, target, ".envrc"))

	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)
	assert.True(t, slices.Contains(config.Migrations, "tk:drop-remote-taskfiles-export"))
}

func TestCoreMigration_deletesAnEnvrcItOwnedAlone(t *testing.T) {
	target := t.TempDir()
	write(t, target, ".envrc", "export TASK_X_REMOTE_TASKFILES=1 #!tk\n")

	_, err := testManager().Write(target, testConfig())
	assert.NoError(t, err)
	assert.False(t, exists(target, ".envrc"))
}

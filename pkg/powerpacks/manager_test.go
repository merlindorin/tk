package powerpacks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

func testManager() *ps.Manager {
	manager := ps.NewPowerpackManager()
	manager.Add(&ps.Powerpack{
		Name:        "git",
		Description: "",
		Taskfile:    []byte("version: '3'\ntasks:\n  hook: {}\n"),
		Readme:      []byte("# Git\n"),
	})
	manager.Add(&ps.Powerpack{
		Name:        "golangci",
		Description: "",
		Taskfile:    []byte("version: '3'\ntasks:\n  lint: {}\n"),
		Readme:      []byte("# Golangci\n"),
	})

	return manager
}

func testConfig() ps.Config {
	return ps.Config{
		Version:        "test",
		IgnoreReadme:   false,
		IgnoreTaskfile: false,
		Includes:       nil,
		Excludes:       nil,
		Powerpacks:     nil,
	}
}

func read(t *testing.T, elem ...string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(elem...))
	assert.NoError(t, err)

	return string(content)
}

func exists(target string, elem ...string) bool {
	_, err := os.Stat(filepath.Join(append([]string{target}, elem...)...))

	return err == nil
}

func TestManagerWrite_onEmptyProject(t *testing.T) {
	target := t.TempDir()

	plan, err := testManager().Write(target, testConfig())
	assert.NoError(t, err)
	assert.Equal(t, len(plan.Changes), len(plan.Pending()))

	assert.Equal(t, "version: '3'\ntasks:\n  hook: {}\n", read(t, target, ".tk", "git", "Taskfile.yaml"))
	assert.Equal(t, "# Git\n", read(t, target, ".tk", "git", "README.md"))
	assert.Contains(t, read(t, target, "Taskfile.yaml"), "git: .tk/git/Taskfile.yaml #!tk")
	assert.False(t, exists(target, ".envrc"), "tk has nothing to write there")
	assert.Contains(t, read(t, target, ".tk.yaml"), "version: test")
}

func TestManagerWrite_isIdempotent(t *testing.T) {
	target := t.TempDir()
	manager := testManager()

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)

	before := read(t, target, "Taskfile.yaml")

	plan, err := manager.Write(target, testConfig())
	assert.NoError(t, err)
	assert.True(t, plan.IsUpToDate())
	assert.Equal(t, before, read(t, target, "Taskfile.yaml"))
}

func TestManagerWrite_keepsUserContent(t *testing.T) {
	target := t.TempDir()
	taskfile := "version: '3'\n\nincludes:\n  docker: ./build/Taskfile.yaml\n\n" +
		"tasks:\n  hello:\n    cmds:\n      - echo hi\n"

	assert.NoError(t, os.WriteFile(filepath.Join(target, "Taskfile.yaml"), []byte(taskfile), 0o600))
	assert.NoError(t, os.WriteFile(filepath.Join(target, ".envrc"), []byte("use flake\n"), 0o600))

	_, err := testManager().Write(target, testConfig())
	assert.NoError(t, err)

	merged := read(t, target, "Taskfile.yaml")
	assert.Contains(t, merged, "docker: ./build/Taskfile.yaml")
	assert.Contains(t, merged, "hello:")
	assert.Contains(t, merged, "git: .tk/git/Taskfile.yaml #!tk")
	assert.Equal(t, "use flake\n", read(t, target, ".envrc"), "an envrc of their own is untouched")
}

func TestManagerWrite_removesExcludedPowerpack(t *testing.T) {
	target := t.TempDir()
	manager := testManager()

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)

	config := testConfig()
	config.Includes = []string{"git"}

	plan, err := manager.Write(target, config)
	assert.NoError(t, err)
	assert.False(t, plan.IsUpToDate())

	assert.False(t, exists(target, ".tk", "golangci"))
	assert.True(t, exists(target, ".tk", "git"))
	assert.NotContains(t, read(t, target, "Taskfile.yaml"), "golangci")
	assert.Contains(t, read(t, target, "Taskfile.yaml"), "git:")
}

func TestManagerPlan_writesNothing(t *testing.T) {
	target := t.TempDir()

	plan, err := testManager().Plan(target, testConfig())
	assert.NoError(t, err)
	assert.True(t, len(plan.Pending()) > 0)

	entries, err := os.ReadDir(target)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(entries))
}

func TestManagerWrite_honoursIgnoreOptions(t *testing.T) {
	target := t.TempDir()
	config := testConfig()
	config.IgnoreReadme = true

	_, err := testManager().Write(target, config)
	assert.NoError(t, err)

	assert.False(t, exists(target, ".tk", "git", "README.md"))
	assert.True(t, exists(target, ".tk", "git", "Taskfile.yaml"))
}

func TestManagerWrite_ignoreTaskfileLeavesRootAlone(t *testing.T) {
	target := t.TempDir()
	config := testConfig()
	config.IgnoreTaskfile = true

	_, err := testManager().Write(target, config)
	assert.NoError(t, err)

	assert.False(t, exists(target, "Taskfile.yaml"))
	assert.True(t, exists(target, ".tk", "git", "README.md"))
}

func TestManagerWrite_recordsProvenance(t *testing.T) {
	target := t.TempDir()

	plan, err := testManager().Write(target, testConfig())
	assert.NoError(t, err)

	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)
	assert.Equal(t, "test", config.Version)
	assert.Equal(t, plan.Config.Powerpacks, config.Powerpacks)

	powerpack, ok := testManager().Get("git")
	assert.True(t, ok)
	assert.Equal(t, powerpack.Checksum(), config.Powerpacks["git"])
}

func TestManagerValidate(t *testing.T) {
	manager := testManager()

	assert.NoError(t, manager.Validate("git", "golangci"))
	assert.IsError(t, manager.Validate("git", "nope"), ps.ErrUnknownPowerpack)
}

func TestManagerNames(t *testing.T) {
	assert.Equal(t, []string{"git", "golangci"}, testManager().Names())
}

func TestManagerSelected(t *testing.T) {
	config := testConfig()
	assert.Equal(t, 2, len(testManager().Selected(config)), "no selection installs everything")

	config.Includes = []string{"golangci"}

	selected := testManager().Selected(config)
	assert.Equal(t, 1, len(selected))
	assert.Equal(t, "golangci", selected[0].Name)
}

func TestManagerEnableAndDisable(t *testing.T) {
	manager := testManager()
	config := testConfig()

	assert.Zero(t, len(manager.Enable(&config, "git")), "everything is installed already")

	removed, err := manager.Disable(&config, "git")
	assert.NoError(t, err)
	assert.Equal(t, []string{"git"}, removed)
	assert.Equal(t, []string{"golangci"}, config.Includes, "the implicit selection is written out")
	assert.False(t, config.IsSelected("git"))

	assert.Equal(t, []string{"git"}, manager.Enable(&config, "git"))
	assert.Equal(t, []string{"git", "golangci"}, config.Includes)
}

func TestManagerDisable_refusesToEmptyTheProject(t *testing.T) {
	manager := testManager()
	config := testConfig()

	_, err := manager.Disable(&config, "git", "golangci")
	assert.IsError(t, err, ps.ErrEmptySelection)
}

func TestManagerPlan_migratesExcludesToIncludes(t *testing.T) {
	target := t.TempDir()
	manager := testManager()

	config := testConfig()
	config.Excludes = []string{"golangci"}

	_, err := manager.Write(target, config)
	assert.NoError(t, err)

	written, err := ps.LoadConfig(target)
	assert.NoError(t, err)
	assert.Equal(t, []string{"git"}, written.Includes)
	assert.Zero(t, len(written.Excludes), "the deny list is dropped once converted")
	assert.False(t, exists(target, ".tk", "golangci"))
}

func TestManagerWrite_usesTheExistingTaskfileName(t *testing.T) {
	// Creating Taskfile.yaml next to a Taskfile.yml would hide the includes from Task.
	target := t.TempDir()
	existing := filepath.Join(target, "Taskfile.yml")

	assert.NoError(t, os.WriteFile(existing, []byte("version: '3'\n\nincludes:\n"), 0o600))

	_, err := testManager().Write(target, testConfig())
	assert.NoError(t, err)

	assert.False(t, exists(target, "Taskfile.yaml"))
	assert.Contains(t, read(t, existing), "git: .tk/git/Taskfile.yaml #!tk")
}

func TestManagerWrite_keepsWindowsLineEndings(t *testing.T) {
	target := t.TempDir()
	taskfile := "version: '3'\r\n\r\nincludes:\r\n  docker: ./build/Taskfile.yaml\r\n\r\ntasks:\r\n  hello: {}\r\n"

	assert.NoError(t, os.WriteFile(filepath.Join(target, "Taskfile.yaml"), []byte(taskfile), 0o600))

	manager := testManager()

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)

	merged := read(t, target, "Taskfile.yaml")
	assert.Contains(t, merged, "git: .tk/git/Taskfile.yaml #!tk\r\n")
	assert.Contains(t, merged, "docker: ./build/Taskfile.yaml\r\n")
	assert.False(t, strings.Contains(strings.ReplaceAll(merged, "\r\n", ""), "\n"), "mixed line endings")

	// And a second run must recognise its own markers rather than duplicate them.
	plan, err := manager.Write(target, testConfig())
	assert.NoError(t, err)
	assert.True(t, plan.IsUpToDate())
}

func TestFindTaskfile(t *testing.T) {
	target := t.TempDir()
	assert.Equal(t, "Taskfile.yaml", must(ps.FindTaskfile(target)), "defaults when the project has none")

	assert.NoError(t, os.WriteFile(filepath.Join(target, "Taskfile.dist.yaml"), []byte("version: '3'\n"), 0o600))
	assert.Equal(t, "Taskfile.dist.yaml", must(ps.FindTaskfile(target)))

	assert.NoError(t, os.WriteFile(filepath.Join(target, "Taskfile.yaml"), []byte("version: '3'\n"), 0o600))
	assert.Equal(t, "Taskfile.yaml", must(ps.FindTaskfile(target)))

	assert.NoError(t, os.WriteFile(filepath.Join(target, "Taskfile.yml"), []byte("version: '3'\n"), 0o600))
	assert.Equal(t, "Taskfile.yml", must(ps.FindTaskfile(target)), "Task prefers Taskfile.yml")
}

func TestFindTaskfile_ignoresDirectories(t *testing.T) {
	target := t.TempDir()
	assert.NoError(t, os.Mkdir(filepath.Join(target, "Taskfile.yml"), 0o750))
	assert.Equal(t, "Taskfile.yaml", must(ps.FindTaskfile(target)))
}

func must(name string, err error) string {
	if err != nil {
		panic(err)
	}

	return name
}

func TestManagerWrite_removesTheEnvrcItUsedToWrite(t *testing.T) {
	// Task released the REMOTE_TASKFILES experiment: the export tk wrote there now makes
	// Task print a warning on every invocation, so an update has to clean it up.
	target := t.TempDir()
	envrc := filepath.Join(target, ".envrc")

	assert.NoError(t, os.WriteFile(envrc, []byte("export TASK_X_REMOTE_TASKFILES=1 #!tk\n"), 0o600))

	_, err := testManager().Write(target, testConfig())
	assert.NoError(t, err)
	assert.False(t, exists(target, ".envrc"), "nothing else was in it")
}

func TestManagerWrite_keepsAnEnvrcTheUserAlsoUses(t *testing.T) {
	target := t.TempDir()
	envrc := filepath.Join(target, ".envrc")
	content := "use flake\nexport TASK_X_REMOTE_TASKFILES=1 #!tk\nexport FOO=bar\n"

	assert.NoError(t, os.WriteFile(envrc, []byte(content), 0o600))

	manager := testManager()

	_, err := manager.Write(target, testConfig())
	assert.NoError(t, err)
	assert.Equal(t, "use flake\nexport FOO=bar\n", read(t, envrc))

	plan, err := manager.Write(target, testConfig())
	assert.NoError(t, err)
	assert.True(t, plan.IsUpToDate(), "the cleanup happens once")
}

package powerpacks_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

func TestLoadConfig_missing(t *testing.T) {
	_, err := ps.LoadConfig(t.TempDir())
	assert.IsError(t, err, ps.ErrConfigNotFound)
}

func TestLoadConfig_fromAPreviousVersion(t *testing.T) {
	// Configurations written before tk recorded provenance, or before the excludes deny
	// list became the includes allow list, must still load and mean the same thing.
	target := t.TempDir()
	legacy := "ignore_readme: false\nignore_taskfile: false\nexcludes:\n    - claude\n"

	assert.NoError(t, os.WriteFile(filepath.Join(target, ".tk.yaml"), []byte(legacy), 0o600))

	config, err := ps.LoadConfig(target)
	assert.NoError(t, err)
	assert.False(t, config.IsSelected("claude"))
	assert.True(t, config.IsSelected("git"))
	assert.False(t, config.SelectsAll())
	assert.Equal(t, "", config.Version)
	assert.Zero(t, len(config.Powerpacks))
}

func TestConfigIsSelected(t *testing.T) {
	empty := ps.Config{
		Version: "", IgnoreReadme: false, IgnoreTaskfile: false, IgnoreEnvrc: false,
		Includes: nil, Excludes: nil, Powerpacks: nil,
	}
	assert.True(t, empty.IsSelected("anything"), "no selection installs everything")
	assert.True(t, empty.SelectsAll())

	named := ps.Config{
		Version: "", IgnoreReadme: false, IgnoreTaskfile: false, IgnoreEnvrc: false,
		Includes: []string{"git"}, Excludes: nil, Powerpacks: nil,
	}
	assert.True(t, named.IsSelected("git"))
	assert.False(t, named.IsSelected("golangci"), "a named selection installs only those")
	assert.False(t, named.SelectsAll())
}

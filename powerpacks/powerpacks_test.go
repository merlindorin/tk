package powerpacks_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/merlindorin/tk/powerpacks"
)

func TestBuildPowerpackManager(t *testing.T) {
	manager, err := powerpacks.BuildPowerpackManager()
	assert.NoError(t, err)

	names := manager.Names()
	assert.True(t, slices.Contains(names, "golangci"))
	assert.True(t, slices.Contains(names, "git"))

	for _, name := range names {
		powerpack, ok := manager.Get(name)
		assert.True(t, ok)
		assert.True(t, powerpack.HasTaskfile(), "%s ships no Taskfile", name)
		assert.True(t, powerpack.HasReadme(), "%s ships no README", name)
		assert.NotZero(t, powerpack.Summary())
	}
}

func TestBuildPowerpackManager_doesNotEmbedSources(t *testing.T) {
	manager, err := powerpacks.BuildPowerpackManager()
	assert.NoError(t, err)

	for _, name := range manager.Names() {
		assert.False(t, strings.HasSuffix(name, ".go"), "%s should not be a powerpack", name)
	}
}

func TestBuildPowerpackManager_isStable(t *testing.T) {
	// Powerpack content is read into memory, so building twice yields the same checksums.
	first, err := powerpacks.BuildPowerpackManager()
	assert.NoError(t, err)

	second, err := powerpacks.BuildPowerpackManager()
	assert.NoError(t, err)

	for _, name := range first.Names() {
		a, ok := first.Get(name)
		assert.True(t, ok)

		b, ok := second.Get(name)
		assert.True(t, ok)
		assert.Equal(t, a.Checksum(), b.Checksum())
	}
}

func TestBuildPowerpackManager_manifestsAreValid(t *testing.T) {
	// A manifest tk cannot honour must fail at load time, not halfway through a project.
	manager, err := powerpacks.BuildPowerpackManager()
	assert.NoError(t, err)

	for _, name := range manager.Names() {
		powerpack, ok := manager.Get(name)
		assert.True(t, ok)

		for _, owned := range powerpack.Owns() {
			assert.False(t, strings.HasPrefix(owned.Path, ".tk/"), "%s claims %s", name, owned.Path)

			if owned.Source != "" {
				assert.True(t, len(powerpack.Sources[owned.Source]) > 0, "%s ships no %s", name, owned.Source)
			}
		}

		for _, required := range powerpack.Requires() {
			_, known := manager.Get(required)
			assert.True(t, known, "%s requires unknown powerpack %s", name, required)
		}
	}
}

func TestBuildPowerpackManager_describesEveryPowerpack(t *testing.T) {
	manager, err := powerpacks.BuildPowerpackManager()
	assert.NoError(t, err)

	for _, powerpack := range manager.List() {
		assert.NotZero(t, powerpack.Manifest.Description, "%s has no manifest description", powerpack.Name)
	}
}

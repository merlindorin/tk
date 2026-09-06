package powerpacks_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

func TestParseManifest(t *testing.T) {
	manifest, err := ps.ParseManifest("git", []byte(`
description: Git pre-commit hooks
requires: [go]
owns:
  - path: .gitignore
  - path: .github/workflows/git.yml
    strategy: sync
    source: files/workflow.yml
migrations:
  - id: drop-old-hook
    reason: the hook moved
    remove-file: .git/hooks/tk
`))
	assert.NoError(t, err)
	assert.Equal(t, "Git pre-commit hooks", manifest.Description)
	assert.Equal(t, []string{"go"}, manifest.Requires)
	assert.Equal(t, 2, len(manifest.Owns))
	assert.Equal(t, ps.StrategySync, manifest.Owns[1].Strategy)
	assert.Equal(t, "drop-old-hook", manifest.Migrations[0].ID)
}

func TestParseManifest_empty(t *testing.T) {
	// A powerpack without a manifest owns nothing and needs nothing.
	manifest, err := ps.ParseManifest("git", nil)
	assert.NoError(t, err)
	assert.Zero(t, len(manifest.Owns))
	assert.Zero(t, len(manifest.Migrations))
}

func TestParseManifest_rejects(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
	}{
		{
			name:     "an unknown strategy",
			manifest: "owns:\n  - path: a.yml\n    strategy: clobber\n",
		},
		{
			name:     "a synced file without a source",
			manifest: "owns:\n  - path: a.yml\n    strategy: sync\n",
		},
		{
			name:     "a source on a file tk does not write",
			manifest: "owns:\n  - path: a.yml\n    source: files/a.yml\n",
		},
		{
			name:     "a path escaping the project",
			manifest: "owns:\n  - path: ../../etc/passwd\n",
		},
		{
			name:     "an absolute path",
			manifest: "owns:\n  - path: /etc/passwd\n",
		},
		{
			name:     "a claim on the directory tk owns",
			manifest: "owns:\n  - path: .tk/git/Taskfile.yaml\n",
		},
		{
			name:     "a migration without an id",
			manifest: "migrations:\n  - remove-file: a.yml\n",
		},
		{
			name:     "a migration doing nothing",
			manifest: "migrations:\n  - id: noop\n",
		},
		{
			name:     "a migration doing two things",
			manifest: "migrations:\n  - id: two\n    remove-file: a.yml\n    rename-file: {from: b.yml, to: c.yml}\n",
		},
		{
			name:     "a migration escaping the project",
			manifest: "migrations:\n  - id: escape\n    remove-file: ../secrets\n",
		},
		{
			name:     "a migration with a broken pattern",
			manifest: "migrations:\n  - id: bad\n    remove-lines: {path: .envrc, matching: '('}\n",
		},
		{
			name:     "two migrations sharing an id",
			manifest: "migrations:\n  - id: same\n    remove-file: a.yml\n  - id: same\n    remove-file: b.yml\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ps.ParseManifest("git", []byte(tt.manifest))
			assert.IsError(t, err, ps.ErrInvalidManifest)
		})
	}
}

package powerpacks

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// migration is a Migration together with the powerpack that declared it, so a report can
// say where an instruction came from.
type migration struct {
	Migration

	owner string
}

// qualifiedID is how an applied migration is recorded in the configuration. Prefixing it
// with the powerpack keeps two packs from colliding on a common id like `drop-config`.
func (m *migration) qualifiedID() string {
	return m.owner + ":" + m.ID
}

// pending returns the migrations of the selected powerpacks, plus the ones tk itself
// carries, that this project has not run yet.
func (m *Manager) pending(config Config) []migration {
	applied := make(map[string]bool, len(config.Migrations))
	for _, id := range config.Migrations {
		applied[id] = true
	}

	declared := coreMigrations()

	for _, powerpack := range m.Selected(config) {
		for _, declaration := range powerpack.Manifest.Migrations {
			declared = append(declared, migration{Migration: declaration, owner: powerpack.Name})
		}
	}

	pending := make([]migration, 0, len(declared))

	for _, candidate := range declared {
		if !applied[candidate.qualifiedID()] {
			pending = append(pending, candidate)
		}
	}

	return pending
}

// migrate runs the pending migrations against the project, returning the files they
// change and the ids to record. Nothing is written: the files join the plan, so a
// migration shows up in a dry run exactly like any other change.
func migrate(target string, pending []migration) ([]file, []string, error) {
	files := map[string]file{}
	ids := make([]string, 0, len(pending))

	for i := range pending {
		changed, err := pending[i].apply(target, files)
		if err != nil {
			return nil, nil, err
		}

		// A migration is recorded whether or not it found anything to change: it has run,
		// and a project that never had the old content must not keep looking for it.
		ids = append(ids, pending[i].qualifiedID())

		for _, f := range changed {
			files[f.path] = f
		}
	}

	return sortedFiles(files), ids, nil
}

// apply computes what a single migration changes, reading through the files earlier
// migrations already changed.
func (m *migration) apply(target string, changed map[string]file) ([]file, error) {
	switch {
	case m.RemoveLines != nil:
		return m.removeLines(target, changed)
	case m.RemoveFile != "":
		return removeFile(target, m.RemoveFile, changed)
	case m.RenameFile != nil:
		return m.renameFile(target, changed)
	default:
		return nil, fmt.Errorf("%w: %s: migration %q does nothing", ErrInvalidManifest, m.owner, m.ID)
	}
}

// removeLines drops the lines matching the pattern, and the file with them when it is
// asked to and nothing but blank lines is left.
func (m *migration) removeLines(target string, changed map[string]file) ([]file, error) {
	current, found, err := readThrough(target, m.RemoveLines.Path, changed)
	if err != nil || !found {
		return nil, err
	}

	pattern, err := regexp.Compile(m.RemoveLines.Matching)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s: %w", ErrInvalidManifest, m.owner, m.ID, err)
	}

	lines, ending := splitLines(string(current))
	kept := make([]string, 0, len(lines))

	for _, line := range lines {
		if !pattern.MatchString(line) {
			kept = append(kept, line)
		}
	}

	if len(kept) == len(lines) {
		return nil, nil
	}

	if m.RemoveLines.DeleteWhenEmpty && strings.TrimSpace(strings.Join(kept, "")) == "" {
		return []file{{path: m.RemoveLines.Path, content: nil, absent: true}}, nil
	}

	return []file{{path: m.RemoveLines.Path, content: []byte(joinLines(kept, ending)), absent: false}}, nil
}

// renameFile moves a file, leaving an existing destination alone: the project has
// already been migrated, or the user wrote something there themselves.
func (m *migration) renameFile(target string, changed map[string]file) ([]file, error) {
	current, found, err := readThrough(target, m.RenameFile.From, changed)
	if err != nil || !found {
		return nil, err
	}

	_, taken, err := readThrough(target, m.RenameFile.To, changed)
	if err != nil || taken {
		return nil, err
	}

	return []file{
		{path: m.RenameFile.From, content: nil, absent: true},
		{path: m.RenameFile.To, content: current, absent: false},
	}, nil
}

// removeFile deletes a file a powerpack used to generate.
func removeFile(target, name string, changed map[string]file) ([]file, error) {
	_, found, err := readThrough(target, name, changed)
	if err != nil || !found {
		return nil, err
	}

	return []file{{path: name, content: nil, absent: true}}, nil
}

// readThrough reads a file as the migrations run so far have left it.
func readThrough(target, name string, changed map[string]file) ([]byte, bool, error) {
	if staged, ok := changed[name]; ok {
		return staged.content, !staged.absent, nil
	}

	return readFile(filepath.Join(target, filepath.FromSlash(name)))
}

package powerpacks

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// ManifestFilename is the optional file where a powerpack declares what it needs, what
// it owns in the project, and how to clean up after itself.
const ManifestFilename = "powerpack.yaml"

// ErrInvalidManifest is reported when a powerpack declares something tk cannot honour.
var ErrInvalidManifest = errors.New("invalid powerpack manifest")

// Strategy tells who writes an owned file and what tk may do with it.
type Strategy string

const (
	// StrategyOnce is for files the powerpack tasks generate: tk never writes them and
	// never deletes them, it only reports the ones left behind by a removed powerpack.
	StrategyOnce Strategy = "once"
	// StrategySync is for files tk writes itself from the powerpack: it keeps them in
	// step with the binary and removes them with the powerpack.
	StrategySync Strategy = "sync"
)

// Manifest is the declaration a powerpack ships in powerpack.yaml. Every field is
// optional: a powerpack without a manifest is a powerpack that owns nothing outside
// `.tk/` and needs nothing.
type Manifest struct {
	Description string      `yaml:"description,omitempty"`
	Requires    []string    `yaml:"requires,omitempty"`
	Owns        []Owned     `yaml:"owns,omitempty"`
	Migrations  []Migration `yaml:"migrations,omitempty"`
}

// Owned is a file a powerpack is responsible for outside of `.tk/`.
type Owned struct {
	Path     string   `yaml:"path"`
	Strategy Strategy `yaml:"strategy,omitempty"`
	Source   string   `yaml:"source,omitempty"`
}

// Strategy of an owned file, defaulting to the conservative one.
func (o *Owned) strategy() Strategy {
	if o.Strategy == "" {
		return StrategyOnce
	}

	return o.Strategy
}

// Migration undoes something a previous version of a powerpack did. It runs once per
// project and is then recorded by id, so it stays cheap forever after.
type Migration struct {
	ID     string `yaml:"id"`
	Reason string `yaml:"reason,omitempty"`

	RemoveLines *RemoveLines `yaml:"remove-lines,omitempty"`
	RemoveFile  string       `yaml:"remove-file,omitempty"`
	RenameFile  *RenameFile  `yaml:"rename-file,omitempty"`
}

// RemoveLines drops the lines of a file matching a pattern, and the file itself when
// asked to and nothing is left in it.
type RemoveLines struct {
	Path            string `yaml:"path"`
	Matching        string `yaml:"matching"`
	DeleteWhenEmpty bool   `yaml:"delete-when-empty,omitempty"`
}

// RenameFile moves a file a powerpack used to generate under a different name.
type RenameFile struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// ParseManifest reads a powerpack manifest and checks that tk can honour it.
func ParseManifest(name string, content []byte) (Manifest, error) {
	manifest := Manifest{} //nolint:exhaustruct // everything in a manifest is optional

	if len(content) == 0 {
		return manifest, nil
	}

	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return manifest, fmt.Errorf("%w: %s: %w", ErrInvalidManifest, name, err)
	}

	if err := manifest.validate(name); err != nil {
		return manifest, err
	}

	return manifest, nil
}

// validate rejects a manifest tk cannot apply safely, so a broken powerpack fails at
// load time rather than halfway through writing a project.
func (m *Manifest) validate(name string) error {
	for i := range m.Owns {
		if err := m.Owns[i].validate(name); err != nil {
			return err
		}
	}

	seen := make(map[string]bool, len(m.Migrations))

	for i := range m.Migrations {
		if err := m.Migrations[i].validate(name); err != nil {
			return err
		}

		if seen[m.Migrations[i].ID] {
			return fmt.Errorf("%w: %s: duplicate migration %q", ErrInvalidManifest, name, m.Migrations[i].ID)
		}

		seen[m.Migrations[i].ID] = true
	}

	return nil
}

// validate checks a single owned file.
func (o *Owned) validate(name string) error {
	if err := validPath(name, "owns", o.Path); err != nil {
		return err
	}

	switch o.strategy() {
	case StrategyOnce:
		if o.Source != "" {
			return fmt.Errorf("%w: %s: %q is written by a task, so it cannot have a source",
				ErrInvalidManifest, name, o.Path)
		}
	case StrategySync:
		if o.Source == "" {
			return fmt.Errorf("%w: %s: %q is written by tk, so it needs a source",
				ErrInvalidManifest, name, o.Path)
		}

		if err := validPath(name, "source", o.Source); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: %s: unknown strategy %q for %q (want %q or %q)",
			ErrInvalidManifest, name, o.Strategy, o.Path, StrategyOnce, StrategySync)
	}

	return nil
}

// validate checks that a migration carries exactly one instruction tk understands.
func (m *Migration) validate(name string) error {
	if m.ID == "" {
		return fmt.Errorf("%w: %s: a migration needs an id", ErrInvalidManifest, name)
	}

	verbs := 0

	if m.RemoveLines != nil {
		verbs++

		if err := validPath(name, m.ID, m.RemoveLines.Path); err != nil {
			return err
		}

		if _, err := regexp.Compile(m.RemoveLines.Matching); err != nil {
			return fmt.Errorf("%w: %s: %s: %w", ErrInvalidManifest, name, m.ID, err)
		}
	}

	if m.RemoveFile != "" {
		verbs++

		if err := validPath(name, m.ID, m.RemoveFile); err != nil {
			return err
		}
	}

	if m.RenameFile != nil {
		verbs++

		if err := validPath(name, m.ID, m.RenameFile.From); err != nil {
			return err
		}

		if err := validPath(name, m.ID, m.RenameFile.To); err != nil {
			return err
		}
	}

	if verbs != 1 {
		return fmt.Errorf("%w: %s: migration %q must carry exactly one instruction, found %d",
			ErrInvalidManifest, name, m.ID, verbs)
	}

	return nil
}

// validPath keeps a manifest from reaching outside the project it is applied to.
func validPath(name, field, value string) error {
	switch {
	case value == "":
		return fmt.Errorf("%w: %s: %s needs a path", ErrInvalidManifest, name, field)
	case path.IsAbs(value), strings.HasPrefix(value, "~"):
		return fmt.Errorf("%w: %s: %s must stay inside the project, got %q", ErrInvalidManifest, name, field, value)
	case slices.Contains(strings.Split(value, "/"), ".."):
		return fmt.Errorf("%w: %s: %s must stay inside the project, got %q", ErrInvalidManifest, name, field, value)
	case strings.HasPrefix(value, PowerpackDir+"/"):
		return fmt.Errorf("%w: %s: %s cannot claim %q, tk owns that directory",
			ErrInvalidManifest, name, field, value)
	}

	return nil
}

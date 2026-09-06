package powerpacks

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Errors reported when a powerpack does not carry the file being asked for.
var (
	ErrMissingTaskfile = errors.New("missing taskfile for powerpack")
	ErrMissingReadme   = errors.New("missing readme for powerpack")
)

// Powerpack is a bundle of reusable configuration: a Taskfile and its documentation.
// Both are held in memory so a powerpack can be written more than once.
type Powerpack struct {
	Name        string
	Description string
	Readme      []byte
	Taskfile    []byte

	// Manifest is what the powerpack declares in powerpack.yaml: what it needs, what it
	// owns in the project, and how to clean up after an older version of itself.
	Manifest Manifest

	// Sources are the files the manifest writes with the sync strategy, by source path.
	Sources map[string][]byte
}

// HasTaskfile reports whether the powerpack carries a Taskfile.
func (p *Powerpack) HasTaskfile() bool {
	return len(p.Taskfile) > 0
}

// HasReadme reports whether the powerpack carries a README.
func (p *Powerpack) HasReadme() bool {
	return len(p.Readme) > 0
}

// WriteTaskfile writes the powerpack Taskfile, reporting ErrMissingTaskfile when it has none.
func (p *Powerpack) WriteTaskfile(writer io.Writer) error {
	if !p.HasTaskfile() {
		return fmt.Errorf("%w: %s", ErrMissingTaskfile, p.Name)
	}

	if _, err := io.Copy(writer, bytes.NewReader(p.Taskfile)); err != nil {
		return fmt.Errorf("cannot write taskfile for powerpack %s: %w", p.Name, err)
	}

	return nil
}

// WriteReadme writes the powerpack README, reporting ErrMissingReadme when it has none.
func (p *Powerpack) WriteReadme(writer io.Writer) error {
	if !p.HasReadme() {
		return fmt.Errorf("%w: %s", ErrMissingReadme, p.Name)
	}

	if _, err := io.Copy(writer, bytes.NewReader(p.Readme)); err != nil {
		return fmt.Errorf("cannot write readme for powerpack %s: %w", p.Name, err)
	}

	return nil
}

// Checksum digests the powerpack content. It is recorded in the configuration so tk can
// tell which revision of a powerpack a project was installed with.
func (p *Powerpack) Checksum() string {
	hash := sha256.New()

	declaration, _ := yaml.Marshal(p.Manifest) // a struct of scalars and slices always marshals

	parts := [][]byte{[]byte(p.Name), p.Taskfile, p.Readme, declaration}
	for _, name := range sortedKeys(p.Sources) {
		parts = append(parts, []byte(name), p.Sources[name])
	}

	for _, part := range parts {
		// sha256 never fails to write.
		_, _ = hash.Write(part)
		_, _ = hash.Write([]byte{0})
	}

	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}

// Summary describes the powerpack in `tk list`: what its manifest declares, else the
// title of its documentation, else its name.
func (p *Powerpack) Summary() string {
	if p.Description != "" {
		return p.Description
	}

	if p.Manifest.Description != "" {
		return p.Manifest.Description
	}

	for _, line := range strings.Split(string(p.Readme), "\n") {
		if title, found := strings.CutPrefix(strings.TrimSpace(line), "# "); found {
			return strings.TrimSpace(title)
		}
	}

	return p.Name
}

// Tasks returns the tasks the powerpack defines, sorted. They are namespaced by the
// powerpack name once included, so `git` defining `install` is run as `task git:install`.
func (p *Powerpack) Tasks() []string {
	var taskfile struct {
		Tasks map[string]yaml.Node `yaml:"tasks"`
	}

	if err := yaml.Unmarshal(p.Taskfile, &taskfile); err != nil {
		return nil
	}

	names := make([]string, 0, len(taskfile.Tasks))
	for name := range taskfile.Tasks {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

// Prefix is the namespace the powerpack tasks live under in the root Taskfile.
func (p *Powerpack) Prefix() string {
	return p.Name + ":"
}

// Owns returns the files the powerpack is responsible for outside of `.tk/`.
func (p *Powerpack) Owns() []Owned {
	return p.Manifest.Owns
}

// Requires returns the powerpacks this one needs to work.
func (p *Powerpack) Requires() []string {
	return p.Manifest.Requires
}

// sortedKeys returns the keys of a map in a stable order.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}

package powerpacks

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

// ConfigFilename is the name of the file tk uses to remember how a project was installed.
const ConfigFilename = ".tk.yaml"

// ErrConfigNotFound is reported when a project has no tk configuration yet.
var ErrConfigNotFound = errors.New("no tk configuration found, run `tk init` first")

// Config is the tk configuration persisted in .tk.yaml. Version and Powerpacks are
// provenance: they record which tk build and which powerpack revisions produced `.tk/`.
//
// Includes is an allow list: leaving it empty installs every powerpack, naming some
// installs only those. Excludes is the deny list tk used before; it is still read so an
// existing project keeps working, and the next write converts it into Includes.
type Config struct {
	Version        string            `yaml:"version,omitempty"`
	IgnoreReadme   bool              `yaml:"ignore_readme"`
	IgnoreTaskfile bool              `yaml:"ignore_taskfile"`
	IgnoreEnvrc    bool              `yaml:"ignore_envrc"`
	Includes       []string          `yaml:"includes,omitempty"`
	Excludes       []string          `yaml:"excludes,omitempty"`
	Powerpacks     map[string]string `yaml:"powerpacks,omitempty"`
}

// LoadConfig reads the tk configuration of a project, reporting ErrConfigNotFound when
// the project has never been initialised.
func LoadConfig(target string) (Config, error) {
	config := Config{} //nolint:exhaustruct // zero value is the documented default

	content, err := os.ReadFile(filepath.Join(target, ConfigFilename))
	if errors.Is(err, os.ErrNotExist) {
		return config, fmt.Errorf("%w: %s", ErrConfigNotFound, filepath.Join(target, ConfigFilename))
	}

	if err != nil {
		return config, fmt.Errorf("failed to read %s: %w", ConfigFilename, err)
	}

	if err = yaml.Unmarshal(content, &config); err != nil {
		return config, fmt.Errorf("failed to parse %s: %w", ConfigFilename, err)
	}

	return config, nil
}

// IsSelected reports whether a powerpack is installed in this project. An empty
// selection means every powerpack, so a project that never named any keeps picking up
// the ones a newer tk ships.
func (c *Config) IsSelected(name string) bool {
	if slices.Contains(c.Excludes, name) {
		return false
	}

	return len(c.Includes) == 0 || slices.Contains(c.Includes, name)
}

// SelectsAll reports whether the project installs every powerpack tk ships.
func (c *Config) SelectsAll() bool {
	return len(c.Includes) == 0 && len(c.Excludes) == 0
}

// marshal renders the configuration as it is stored on disk.
func (c *Config) marshal() ([]byte, error) {
	content, err := yaml.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to encode %s: %w", ConfigFilename, err)
	}

	return content, nil
}

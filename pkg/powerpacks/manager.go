package powerpacks

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// Names of the files and directories tk manages in a project.
const (
	// TaskfileFilename is the root Taskfile, shared with the user.
	TaskfileFilename = "Taskfile.yaml"
	// ReadmeFilename is the documentation shipped with a powerpack.
	ReadmeFilename = "README.md"
	// EnvrcFilename is the direnv file, shared with the user.
	EnvrcFilename = ".envrc"
	// PowerpackDir is the directory tk owns entirely.
	PowerpackDir = ".tk"
)

const (
	dirMode  = 0o755
	fileMode = 0o644
)

// taskfileNames are the root Taskfile names Task itself looks for, in its own order of
// preference. tk merges its includes into the one the project already uses, and falls
// back to TaskfileFilename when the project has none.
//
//nolint:gochecknoglobals // the list is the contract of the task runner, not state
var taskfileNames = []string{
	"Taskfile.yml",
	"Taskfile.yaml",
	"taskfile.yml",
	"taskfile.yaml",
	"Taskfile.dist.yml",
	"Taskfile.dist.yaml",
	"taskfile.dist.yml",
	"taskfile.dist.yaml",
}

// Errors reported when a selection cannot be honoured.
var (
	// ErrUnknownPowerpack is reported when a name does not match any embedded powerpack.
	ErrUnknownPowerpack = errors.New("unknown powerpack")
	// ErrEmptySelection is reported when a change would leave the project with none.
	ErrEmptySelection = errors.New("a project must keep at least one powerpack")
)

// Action tells what a Change does to a file.
type Action string

// The operations a Change can describe.
const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionKeep   Action = "unchanged"
)

// Change is a single file operation. Path is relative to the project, slash separated.
type Change struct {
	Path   string
	Action Action

	content []byte
}

// Plan is everything tk would do to bring a project in sync with its powerpacks. It is
// computed without touching the filesystem, so it doubles as the dry run and the drift
// report behind `tk status`.
type Plan struct {
	Target  string
	Config  Config
	Changes []Change
}

// Pending returns the changes that are not already satisfied on disk.
func (p *Plan) Pending() []Change {
	pending := make([]Change, 0, len(p.Changes))

	for _, change := range p.Changes {
		if change.Action != ActionKeep {
			pending = append(pending, change)
		}
	}

	return pending
}

// IsUpToDate reports whether the project already matches the plan.
func (p *Plan) IsUpToDate() bool {
	return len(p.Pending()) == 0
}

// Apply performs the plan.
func (p *Plan) Apply() error {
	for _, change := range p.Changes {
		filename := filepath.Join(p.Target, filepath.FromSlash(change.Path))

		var err error

		switch change.Action {
		case ActionDelete:
			err = os.Remove(filename)
		case ActionCreate, ActionUpdate:
			err = writeFile(filename, change.content)
		case ActionKeep:
			continue
		}

		if err != nil {
			return fmt.Errorf("failed to %s %s: %w", change.Action, change.Path, err)
		}
	}

	return pruneEmptyDirs(filepath.Join(p.Target, PowerpackDir))
}

// Manager holds the powerpacks tk knows about.
type Manager struct {
	powerpacks map[string]*Powerpack
}

// NewPowerpackManager returns an empty Manager.
func NewPowerpackManager() *Manager {
	return &Manager{powerpacks: map[string]*Powerpack{}}
}

// Add registers a powerpack.
func (m *Manager) Add(powerpack *Powerpack) {
	m.powerpacks[powerpack.Name] = powerpack
}

// Del unregisters a powerpack.
func (m *Manager) Del(name string) {
	delete(m.powerpacks, name)
}

// Get returns a powerpack by name.
func (m *Manager) Get(name string) (*Powerpack, bool) {
	powerpack, ok := m.powerpacks[name]

	return powerpack, ok
}

// Names returns every known powerpack name, sorted.
func (m *Manager) Names() []string {
	names := make([]string, 0, len(m.powerpacks))
	for name := range m.powerpacks {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

// List returns every known powerpack, sorted by name.
func (m *Manager) List() []Powerpack {
	list := make([]Powerpack, 0, len(m.powerpacks))
	for _, name := range m.Names() {
		list = append(list, *m.powerpacks[name])
	}

	return list
}

// Selected returns the powerpacks a configuration installs, sorted by name.
func (m *Manager) Selected(config Config) []Powerpack {
	selected := make([]Powerpack, 0, len(m.powerpacks))

	for _, powerpack := range m.List() {
		if config.IsSelected(powerpack.Name) {
			selected = append(selected, powerpack)
		}
	}

	return selected
}

// Enable adds powerpacks to the selection and reports the names it changed. A project
// that installs everything already has them, so there is nothing to change.
func (m *Manager) Enable(config *Config, names ...string) []string {
	if len(config.Includes) == 0 && len(config.Excludes) == 0 {
		return nil
	}

	changed := make([]string, 0, len(names))

	for _, name := range names {
		if config.IsSelected(name) {
			continue
		}

		config.Excludes = slices.DeleteFunc(config.Excludes, func(excluded string) bool { return excluded == name })

		if len(config.Includes) > 0 {
			config.Includes = append(config.Includes, name)
		}

		changed = append(changed, name)
	}

	slices.Sort(config.Includes)

	return changed
}

// Disable removes powerpacks from the selection and reports the names it changed. A
// project that installs everything gets the list written out first, since dropping one
// powerpack is exactly what turns the implicit selection into an explicit one.
func (m *Manager) Disable(config *Config, names ...string) ([]string, error) {
	if len(config.Includes) == 0 {
		config.Includes = powerpackNames(m.Selected(*config))
		config.Excludes = nil
	}

	changed := make([]string, 0, len(names))

	for _, name := range names {
		if !config.IsSelected(name) {
			continue
		}

		config.Includes = slices.DeleteFunc(config.Includes, func(included string) bool { return included == name })
		changed = append(changed, name)
	}

	if len(config.Includes) == 0 {
		return nil, fmt.Errorf("%w: remove %s and %s to uninstall tk instead",
			ErrEmptySelection, ConfigFilename, PowerpackDir)
	}

	return changed, nil
}

// normalize sorts and deduplicates the selection, and turns the excludes deny list of an
// older project into the includes allow list, keeping exactly the powerpacks it had. A
// converted selection stops following new powerpacks, which is what naming them means.
func (m *Manager) normalize(config Config) Config {
	if len(config.Excludes) > 0 && len(config.Includes) == 0 {
		config.Includes = powerpackNames(m.Selected(config))
	}

	config.Excludes = nil

	slices.Sort(config.Includes)

	config.Includes = slices.Compact(config.Includes)

	return config
}

// Validate reports the names that do not match any known powerpack.
func (m *Manager) Validate(names ...string) error {
	unknown := make([]string, 0, len(names))

	for _, name := range names {
		if _, ok := m.powerpacks[name]; !ok {
			unknown = append(unknown, name)
		}
	}

	if len(unknown) > 0 {
		return fmt.Errorf("%w: %s (available: %s)",
			ErrUnknownPowerpack, strings.Join(unknown, ", "), strings.Join(m.Names(), ", "))
	}

	return nil
}

// Plan computes the changes needed in target without writing anything.
func (m *Manager) Plan(target string, config Config) (*Plan, error) {
	config = m.normalize(config)
	selected := m.Selected(config)

	config.Powerpacks = map[string]string{}
	for i := range selected {
		config.Powerpacks[selected[i].Name] = selected[i].Checksum()
	}

	desired, err := desiredFiles(target, config, selected)
	if err != nil {
		return nil, err
	}

	changes, err := diff(target, desired)
	if err != nil {
		return nil, err
	}

	return &Plan{Target: target, Config: config, Changes: changes}, nil
}

// Write plans the changes and applies them, returning what was done.
func (m *Manager) Write(target string, config Config) (*Plan, error) {
	plan, err := m.Plan(target, config)
	if err != nil {
		return nil, err
	}

	if err = plan.Apply(); err != nil {
		return plan, err
	}

	return plan, nil
}

// powerpackNames returns the name of each powerpack of a list.
func powerpackNames(list []Powerpack) []string {
	selected := make([]string, 0, len(list))
	for i := range list {
		selected = append(selected, list[i].Name)
	}

	return selected
}

// file is a file tk is responsible for, with the content it should hold. A file marked
// absent is one tk used to write and no longer does: it is removed from the project.
type file struct {
	path    string
	content []byte
	absent  bool
}

// desiredFiles renders every file tk is responsible for, in a stable order.
func desiredFiles(target string, config Config, selected []Powerpack) ([]file, error) {
	files := make([]file, 0, 2*len(selected)+3) //nolint:mnd // two files per powerpack plus the shared three

	for i := range selected {
		powerpack := &selected[i]

		if !config.IgnoreTaskfile && powerpack.HasTaskfile() {
			files = append(files, file{
				path: powerpackPath(powerpack.Name, TaskfileFilename), content: powerpack.Taskfile, absent: false,
			})
		}

		if !config.IgnoreReadme && powerpack.HasReadme() {
			files = append(files, file{
				path: powerpackPath(powerpack.Name, ReadmeFilename), content: powerpack.Readme, absent: false,
			})
		}
	}

	if !config.IgnoreTaskfile {
		taskfile, err := rootTaskfile(target, selected)
		if err != nil {
			return nil, err
		}

		files = append(files, taskfile)
	}

	if !config.IgnoreEnvrc {
		current, _, err := readFile(filepath.Join(target, EnvrcFilename))
		if err != nil {
			return nil, err
		}

		// tk no longer writes anything to .envrc; what is left is what the user wrote.
		merged := MergeEnvrc(string(current))
		files = append(files, file{path: EnvrcFilename, content: []byte(merged), absent: merged == ""})
	}

	content, err := config.marshal()
	if err != nil {
		return nil, err
	}

	return append(files, file{path: ConfigFilename, content: content, absent: false}), nil
}

// rootTaskfile merges the powerpack includes into the Taskfile the user owns.
func rootTaskfile(target string, selected []Powerpack) (file, error) {
	includes := map[string]string{}

	for i := range selected {
		if selected[i].HasTaskfile() {
			includes[selected[i].Name] = powerpackPath(selected[i].Name, TaskfileFilename)
		}
	}

	name, err := FindTaskfile(target)
	if err != nil {
		return file{}, err //nolint:exhaustruct // the error is what matters
	}

	current, _, err := readFile(filepath.Join(target, name))
	if err != nil {
		return file{}, err //nolint:exhaustruct // the error is what matters
	}

	merged, err := MergeTaskfile(string(current), includes)
	if err != nil {
		return file{}, err //nolint:exhaustruct // the error is what matters
	}

	return file{path: name, content: []byte(merged), absent: false}, nil
}

// FindTaskfile returns the root Taskfile of a project. Writing to the name Task actually
// reads matters: creating Taskfile.yaml next to an existing Taskfile.yml would leave the
// includes in a file the task runner never opens.
func FindTaskfile(target string) (string, error) {
	for _, name := range taskfileNames {
		info, err := os.Stat(filepath.Join(target, name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil {
			return "", fmt.Errorf("failed to inspect %s: %w", name, err)
		}

		if !info.IsDir() {
			return name, nil
		}
	}

	return TaskfileFilename, nil
}

// diff compares the files tk wants with the ones already there.
func diff(target string, desired []file) ([]Change, error) {
	wanted := make(map[string]bool, len(desired))

	for _, f := range desired {
		if !f.absent {
			wanted[f.path] = true
		}
	}

	stale, err := staleFiles(target, wanted)
	if err != nil {
		return nil, err
	}

	changes := make([]Change, 0, len(desired)+len(stale))
	for _, p := range stale {
		changes = append(changes, Change{Path: p, Action: ActionDelete, content: nil})
	}

	for _, f := range desired {
		current, found, er := readFile(filepath.Join(target, filepath.FromSlash(f.path)))
		if er != nil {
			return nil, er
		}

		if f.absent {
			if found {
				changes = append(changes, Change{Path: f.path, Action: ActionDelete, content: nil})
			}

			continue
		}

		action := ActionCreate

		switch {
		case found && bytes.Equal(current, f.content):
			action = ActionKeep
		case found:
			action = ActionUpdate
		}

		changes = append(changes, Change{Path: f.path, Action: action, content: f.content})
	}

	return changes, nil
}

// staleFiles lists the files left in the powerpack directory that tk no longer wants,
// which is how excluded or renamed powerpacks get cleaned up.
func staleFiles(target string, wanted map[string]bool) ([]string, error) {
	root := filepath.Join(target, PowerpackDir)

	var stale []string

	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(target, name)
		if err != nil {
			return fmt.Errorf("failed to resolve %s: %w", name, err)
		}

		if slashed := filepath.ToSlash(rel); !wanted[slashed] {
			stale = append(stale, slashed)
		}

		return nil
	})

	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to inspect %s: %w", root, err)
	}

	slices.Sort(stale)

	return stale, nil
}

// pruneEmptyDirs removes the powerpack directories left empty after a cleanup.
func pruneEmptyDirs(root string) error {
	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to read %s: %w", root, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(root, entry.Name())

		children, er := os.ReadDir(dir)
		if er != nil {
			return fmt.Errorf("failed to read %s: %w", dir, er)
		}

		if len(children) > 0 {
			continue
		}

		if er = os.Remove(dir); er != nil {
			return fmt.Errorf("failed to remove %s: %w", dir, er)
		}
	}

	return nil
}

// powerpackPath is where a powerpack file lives inside the project.
func powerpackPath(name, filename string) string {
	return path.Join(PowerpackDir, name, filename)
}

// readFile reads a file, reporting whether it exists rather than failing on a missing one.
func readFile(filename string) ([]byte, bool, error) {
	content, err := os.ReadFile(filename)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("failed to read %s: %w", filename, err)
	}

	return content, true, nil
}

// writeFile writes a file, creating the directories leading to it.
func writeFile(filename string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(filename), dirMode); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(filename), err)
	}

	if err := os.WriteFile(filename, content, fileMode); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}

	return nil
}

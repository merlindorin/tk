package powerpacks

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

// PowerpackFiles holds the powerpack bundles shipped with the binary. Only the files a
// powerpack is made of are embedded, so the Go sources of this package stay out.
//
//go:embed */Taskfile.yaml */README.md
var PowerpackFiles embed.FS

// BuildPowerpackManager loads every embedded powerpack into a manager.
func BuildPowerpackManager() (*ps.Manager, error) {
	manager := ps.NewPowerpackManager()

	entries, err := fs.ReadDir(PowerpackFiles, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to list powerpacks: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		powerpack, er := readPowerpack(entry.Name())
		if er != nil {
			return nil, er
		}

		manager.Add(powerpack)
	}

	return manager, nil
}

// readPowerpack loads the files of a single embedded powerpack.
func readPowerpack(name string) (*ps.Powerpack, error) {
	powerpack := &ps.Powerpack{ //nolint:exhaustruct // content is filled in below
		Name: name,
	}

	taskfile, err := readIfExist(path.Join(name, ps.TaskfileFilename))
	if err != nil {
		return nil, err
	}

	readme, err := readIfExist(path.Join(name, ps.ReadmeFilename))
	if err != nil {
		return nil, err
	}

	powerpack.Taskfile = taskfile
	powerpack.Readme = readme

	return powerpack, nil
}

// readIfExist reads an embedded file, tolerating a powerpack that ships only one of them.
func readIfExist(name string) ([]byte, error) {
	content, err := PowerpackFiles.ReadFile(name)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", name, err)
	}

	return content, nil
}

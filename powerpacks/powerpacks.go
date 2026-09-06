package powerpacks

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

// PowerpackFiles holds the powerpack bundles shipped with the binary. Only the powerpack
// directories are embedded, so the Go sources of this package stay out.
//
//go:embed */Taskfile.yaml */README.md */powerpack.yaml */files/*
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

	declaration, err := readIfExist(path.Join(name, ps.ManifestFilename))
	if err != nil {
		return nil, err
	}

	manifest, err := ps.ParseManifest(name, declaration)
	if err != nil {
		return nil, err
	}

	sources, err := readSources(name, manifest)
	if err != nil {
		return nil, err
	}

	powerpack.Taskfile = taskfile
	powerpack.Readme = readme
	powerpack.Description = manifest.Description
	powerpack.Manifest = manifest
	powerpack.Sources = sources

	return powerpack, nil
}

// readSources loads the files the manifest writes with the sync strategy, failing loudly
// when a powerpack declares one it does not ship.
func readSources(name string, manifest ps.Manifest) (map[string][]byte, error) {
	sources := map[string][]byte{}

	for _, owned := range manifest.Owns {
		if owned.Source == "" {
			continue
		}

		content, err := PowerpackFiles.ReadFile(path.Join(name, owned.Source))
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %q declares source %q: %w",
				ps.ErrInvalidManifest, name, owned.Path, owned.Source, err)
		}

		sources[owned.Source] = content
	}

	return sources, nil
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

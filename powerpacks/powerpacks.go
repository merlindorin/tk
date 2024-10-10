package powerpacks

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

//go:embed *
var PowerpackFiles embed.FS

func BuildPowerpackManager() (*ps.Manager, error) {
	powerpacks := map[string]*ps.Powerpack{}

	if er := fs.WalkDir(PowerpackFiles, ".", func(path string, _ fs.DirEntry, _ error) error {
		dir := filepath.Dir(path)
		powerpackName := filepath.Base(dir)
		filename := filepath.Base(path)

		if powerpackName == "." {
			return nil
		}

		if powerpacks[powerpackName] == nil {
			powerpacks[powerpackName] = &ps.Powerpack{
				Name: powerpackName,
			}
		}

		fsource, err := PowerpackFiles.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}

		switch filename {
		case "Taskfile.yaml":
			powerpacks[powerpackName].Taskfile = fsource
		case "aqua.yaml":
			powerpacks[powerpackName].Aqua = fsource
		case "README.md":
			powerpacks[powerpackName].Readme = fsource
		}

		return nil
	}); er != nil {
		return nil, er
	}

	m := ps.NewPowerpackManager()
	for _, powerpack := range powerpacks {
		m.Add(powerpack)
	}

	return m, nil
}

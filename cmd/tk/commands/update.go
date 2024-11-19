package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
	"github.com/merlindorin/tk/powerpacks"
	"gopkg.in/yaml.v3"

	"github.com/merlindorin/go-shared/pkg/cmd"
)

type UpdateCmd struct {
	Target string `help:"target where to init tk" default:"."`
}

func (i *UpdateCmd) Run(_ *cmd.Commons) error {
	manager, err := powerpacks.BuildPowerpackManager()
	if err != nil {
		return fmt.Errorf("failed to create powerpacks: %w", err)
	}

	f, err := os.ReadFile(filepath.Join(i.Target, ".tk.yaml"))
	if err != nil {
		return fmt.Errorf("failed to read .tk.yaml: %w", err)
	}

	opts := ps.WriteOption{}

	err = yaml.NewDecoder(bytes.NewBuffer(f)).Decode(&opts)
	if err != nil {
		return fmt.Errorf("failed to parse .tk.yaml: %w", err)
	}

	if er := manager.Write(".", opts); er != nil {
		return fmt.Errorf("failed to updates powerpacks: %w", er)
	}

	return nil
}

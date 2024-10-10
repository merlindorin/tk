package powerpacks

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

var (
	ErrMissingTaskfile = fmt.Errorf("missing taskfile for Powerpack")
	ErrMissingAqua     = fmt.Errorf("missing aqua for Powerpack")
)

type Manager struct {
	powerpacks map[string]*Powerpack
}

func NewPowerpackManager() *Manager {
	return &Manager{
		powerpacks: map[string]*Powerpack{},
	}
}

func (p *Manager) Add(po *Powerpack) {
	p.powerpacks[po.Name] = po
}

func (p *Manager) Del(name string) {
	delete(p.powerpacks, name)
}

func (p *Manager) List(l *[]Powerpack) {
	for _, powerpack := range p.powerpacks {
		*l = append(*l, *powerpack)
	}
}

type WriteOption struct {
	IgnoreReadme   bool
	IgnoreAqua     bool
	IgnoreTaskfile bool
}

func (p *Manager) Write(target string, options WriteOption) error {
	var err error

	envrcTmpl := template.Must(template.New("envrc").Parse(envrcTemplate))
	taskfileTmpl := template.Must(template.New("taskfile").Parse(taskfileTemplate))
	aquaTmpl := template.Must(template.New("aqua").Parse(aquaTemplate))

	file, err := EnsureFileExist(target)
	if err != nil {
		return err
	}

	err = envrcTmpl.Execute(file, nil)
	if err != nil {
		return fmt.Errorf("failed to create envrc: %w", err)
	}

	taskfiles := map[string]string{}
	aquafiles := map[string]string{}

	for name, powerpack := range p.powerpacks {
		err = os.MkdirAll(filepath.Join(target, ".tk", name), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", name, err)
		}

		if !options.IgnoreReadme && powerpack.Readme != nil {
			err = writeReadme(target, name, err, powerpack)
			if err != nil {
				return fmt.Errorf("cannot write %s: %w", name, err)
			}
		}

		if !options.IgnoreTaskfile && powerpack.Taskfile != nil {
			err = listTaskfile(target, name, powerpack, taskfiles)
			if err != nil {
				return err
			}
		}

		if !options.IgnoreAqua && powerpack.Aqua != nil {
			err = listAqua(target, name, powerpack, aquafiles)
			if err != nil {
				return err
			}
		}
	}

	if !options.IgnoreTaskfile {
		err = processTaskfile(target, taskfileTmpl, taskfiles)
		if err != nil {
			return err
		}
	}

	if !options.IgnoreAqua {
		err = processAqua(target, aquaTmpl, aquafiles)
		if err != nil {
			return err
		}
	}

	return err
}

func processAqua(target string, aquaTmpl *template.Template, aquafiles map[string]string) error {
	f, er := os.Create(filepath.Join(target, "aqua.yaml"))
	if er != nil {
		return fmt.Errorf("cannot open aqua: %w", er)
	}

	er = aquaTmpl.Execute(f, aquafiles)
	if er != nil {
		return fmt.Errorf("cannot render template: %w", er)
	}
	return nil
}

func processTaskfile(target string, taskfileTmpl *template.Template, taskfiles map[string]string) error {
	f, er := os.Create(filepath.Join(target, "Taskfile.yaml"))
	if er != nil {
		return fmt.Errorf("cannot open taskfile: %w", er)
	}

	er = taskfileTmpl.Execute(f, taskfiles)
	if er != nil {
		return fmt.Errorf("cannot render template: %w", er)
	}
	return nil
}

func listAqua(target string, name string, powerpack *Powerpack, aquafiles map[string]string) error {
	filename := filepath.Join(target, ".tk", name, "aqua.yaml")

	f, err := os.Create(filename)
	defer func(f *os.File) {
		err = errors.Join(err, f.Close())
	}(f)

	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}

	er := powerpack.WriteAqua(f)
	if er != nil && !errors.Is(er, ErrMissingAqua) {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	if errors.Is(er, ErrMissingAqua) {
		return nil
	}

	if err == nil {
		aquafiles[name] = filename
	}

	return err
}

func listTaskfile(target string, name string, powerpack *Powerpack, taskfiles map[string]string) error {
	filename := filepath.Join(target, ".tk", name, "Taskfile.yaml")
	f, err := os.Create(filename)
	defer func(f *os.File) {
		err = errors.Join(err, f.Close())
	}(f)

	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}

	err = powerpack.WriteTaskfile(f)
	if err != nil && !errors.Is(err, ErrMissingTaskfile) {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	if err == nil {
		taskfiles[name] = filename
	}
	return err
}

func writeReadme(target string, name string, err error, powerpack *Powerpack) error {
	filename := filepath.Join(target, ".tk", name, "README.md")
	f, err := os.Create(filename)
	defer func(f *os.File) {
		err = errors.Join(err, f.Close())
	}(f)

	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}

	err = powerpack.WriteReadme(f)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}
	return err
}

func EnsureFileExist(target string) (*os.File, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", target, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", target)
	}

	if er := os.MkdirAll(filepath.Join(target, ".tk"), os.ModePerm); er != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", ".tk", er)
	}

	file, err := os.Create(filepath.Join(target, ".envrc"))
	if err != nil {
		return nil, fmt.Errorf("cannot open envrc: %w", err)
	}
	return file, nil
}

func (p *Manager) Filter(filters []string) *Manager {
	powerpacks := Manager{}

	for name, powerpack := range p.powerpacks {
		if len(filters) == 0 {
			powerpacks.Add(powerpack)
			continue
		}

		for _, filter := range filters {
			if filter == name {
				powerpacks.Add(powerpack)
			}
		}
	}

	return &powerpacks
}

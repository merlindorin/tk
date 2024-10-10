package powerpacks

import (
	"fmt"
	"io"
)

type Powerpack struct {
	Name        string
	Description string
	Readme      io.Reader
	Taskfile    io.Reader
	Aqua        io.Reader
}

func (p *Powerpack) WriteReadme(writer io.Writer) error {
	if p.Readme == nil {
		return nil
	}

	if _, err := io.Copy(writer, p.Readme); err != nil {
		return fmt.Errorf("cannot write readme for Powerpack %s: %w", p.Name, err)
	}

	return nil
}

func (p *Powerpack) WriteTaskfile(writer io.Writer) error {
	if p.Readme == nil {
		return ErrMissingTaskfile
	}

	if _, err := io.Copy(writer, p.Taskfile); err != nil {
		return fmt.Errorf("cannot write taskfile for Powerpack %s: %w", p.Name, err)
	}

	return nil
}

func (p *Powerpack) WriteAqua(writer io.Writer) error {
	if p.Aqua == nil {
		return ErrMissingAqua
	}

	if _, err := io.Copy(writer, p.Aqua); err != nil {
		return fmt.Errorf("cannot write aqua for Powerpack %s: %w", p.Name, err)
	}

	return nil
}

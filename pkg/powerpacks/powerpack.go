package powerpacks

import (
	"fmt"
	"io"
)

type Powerpack struct {
	Name        string
	Description string
	Target      string
	Readme      io.Reader
	Taskfile    io.Reader
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

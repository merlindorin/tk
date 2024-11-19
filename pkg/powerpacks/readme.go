package powerpacks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const ReadmeRegex = "(?i)^readme(?:\\.md)?$"

type ReadmeProcessor struct {
	Readme    io.Reader
	Powerpack *Powerpack
	Regexp    *regexp.Regexp
	Filename  string

	Writer Writer
}

type Writer func(name string) (io.Writer, error)

func NewReadmeProcessor() *ReadmeProcessor {
	return &ReadmeProcessor{
		Regexp: regexp.MustCompile(ReadmeRegex),
		Readme: strings.NewReader(""),
		Writer: func(name string) (io.Writer, error) {
			f, err := os.Open(name)
			defer func() {
				err = errors.Join(err, f.Close())
			}()

			return f, err
		},
	}
}

func ReadmeProcessorBuilder(p *Powerpack) Processor {
	r := NewReadmeProcessor()
	r.Powerpack = p
	return r
}

// Match check if the file should be considered to be a README file.
func (pr *ReadmeProcessor) Match(p string) bool {
	if pr.Regexp == nil {
		return false
	}

	return pr.Regexp.MatchString(p)
}

func (pr *ReadmeProcessor) Collect(_ context.Context, rel string, r io.Reader) error {
	if !pr.Match(rel) {
		return nil
	}

	pr.Filename = filepath.Base(rel)
	pr.Readme = r

	return nil
}

func (pr *ReadmeProcessor) Write(_ context.Context, p string) error {
	f, err := pr.Writer(filepath.Join(p, pr.Filename))
	if err != nil {
		return fmt.Errorf("cannot open: %w", err)
	}

	_, err = io.Copy(f, pr.Readme)

	return err
}

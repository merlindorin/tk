package powerpacks

import (
	"context"
	"io"
)

type ProcessorBuilder func(p *Powerpack) Processor

type Processor interface {
	Collect(_ context.Context, rel string, r io.Reader) error
	Write(ctx context.Context, path string) error
}

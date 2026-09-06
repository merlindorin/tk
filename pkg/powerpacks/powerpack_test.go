package powerpacks_test

import (
	"bytes"
	"testing"

	"github.com/alecthomas/assert/v2"

	ps "github.com/merlindorin/tk/pkg/powerpacks"
)

func TestPowerpackWriteTaskfile(t *testing.T) {
	powerpack := &ps.Powerpack{
		Name: "git", Description: "",
		Taskfile: []byte("version: '3'\n"), Readme: nil,
	}

	buffer := &bytes.Buffer{}
	assert.NoError(t, powerpack.WriteTaskfile(buffer))
	assert.Equal(t, "version: '3'\n", buffer.String())
}

func TestPowerpackWriteTaskfile_withoutTaskfile(t *testing.T) {
	// A powerpack shipping only documentation must report the missing Taskfile, not panic.
	powerpack := &ps.Powerpack{
		Name: "docs", Description: "",
		Taskfile: nil, Readme: []byte("# Docs\n"),
	}

	assert.IsError(t, powerpack.WriteTaskfile(&bytes.Buffer{}), ps.ErrMissingTaskfile)
}

func TestPowerpackWriteReadme_withoutReadme(t *testing.T) {
	// A powerpack shipping only a Taskfile is valid: it must not be reported as missing one.
	powerpack := &ps.Powerpack{
		Name: "git", Description: "",
		Taskfile: []byte("version: '3'\n"), Readme: nil,
	}

	assert.NoError(t, powerpack.WriteTaskfile(&bytes.Buffer{}))
	assert.IsError(t, powerpack.WriteReadme(&bytes.Buffer{}), ps.ErrMissingReadme)
}

func TestPowerpackChecksum(t *testing.T) {
	powerpack := &ps.Powerpack{
		Name: "git", Description: "",
		Taskfile: []byte("version: '3'\n"), Readme: []byte("# Git\n"),
	}
	other := &ps.Powerpack{
		Name: "git", Description: "",
		Taskfile: []byte("version: '3'\ntasks: {}\n"), Readme: []byte("# Git\n"),
	}

	assert.Equal(t, powerpack.Checksum(), powerpack.Checksum())
	assert.NotEqual(t, powerpack.Checksum(), other.Checksum())
	assert.Contains(t, powerpack.Checksum(), "sha256:")
}

func TestPowerpackSummary(t *testing.T) {
	tests := []struct {
		name      string
		powerpack ps.Powerpack
		want      string
	}{
		{
			name: "should use the readme title",
			powerpack: ps.Powerpack{
				Name: "git", Description: "",
				Taskfile: nil, Readme: []byte("# Taskfile Git Pre-Commit Hook\n\n> hello\n"),
			},
			want: "Taskfile Git Pre-Commit Hook",
		},
		{
			name: "should prefer an explicit description",
			powerpack: ps.Powerpack{
				Name: "git", Description: "explicit",
				Taskfile: nil, Readme: []byte("# Title\n"),
			},
			want: "explicit",
		},
		{
			name: "should fall back to the name",
			powerpack: ps.Powerpack{
				Name: "git", Description: "",
				Taskfile: nil, Readme: []byte("no heading here\n"),
			},
			want: "git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.powerpack.Summary())
		})
	}
}

func TestPowerpackTasksAndPrefix(t *testing.T) {
	powerpack := &ps.Powerpack{
		Name: "git", Description: "",
		Taskfile: []byte("version: '3'\n\ntasks:\n  install: {}\n  default: {}\n"), Readme: nil,
	}

	assert.Equal(t, "git:", powerpack.Prefix())
	assert.Equal(t, []string{"default", "install"}, powerpack.Tasks(), "sorted, so the listing is stable")
}

func TestPowerpackTasks_withoutAValidTaskfile(t *testing.T) {
	powerpack := &ps.Powerpack{
		Name: "git", Description: "",
		Taskfile: []byte("\tnot yaml at all"), Readme: nil,
	}

	assert.Zero(t, len(powerpack.Tasks()))
}

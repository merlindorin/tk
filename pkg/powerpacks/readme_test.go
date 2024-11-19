package powerpacks_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/merlindorin/tk/pkg/powerpacks"
)

func TestReadmeProcessor_Match(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should Match README.md",
			args: args{
				p: "README.md",
			},
			want: true,
		},
		{
			name: "should Match README.md",
			args: args{
				p: "README.md",
			},
			want: true,
		},
		{
			name: "should Match readme.md",
			args: args{
				p: "readme.md",
			},
			want: true,
		},
		{
			name: "should Match Readme.md",
			args: args{
				p: "Readme.md",
			},
			want: true,
		},
		{
			name: "should Match README",
			args: args{
				p: "README",
			},
			want: true,
		},
		{
			name: "should Match Readme",
			args: args{
				p: "Readme",
			},
			want: true,
		},
		{
			name: "should not Match readme.markdown",
			args: args{
				p: "readme.markdown",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := powerpacks.NewReadmeProcessor()

			if got := pr.Match(tt.args.p); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadmeProcessor_Collect(t *testing.T) {
	type args struct {
		rel string
		r   io.Reader
	}
	tests := []struct {
		name    string
		args    []args
		wantErr bool
		want    string
	}{
		{
			name: "should not collect a file that is not a readme",
			args: []args{
				{
					rel: "not-a-readme",
					r:   strings.NewReader("not-a-readme-content"),
				},
			},
			want: "",
		},
		{
			name: "should collect a file that is a readme",
			args: []args{
				{
					rel: "Readme.md",
					r:   strings.NewReader("a-readme-content"),
				},
			},
			want: "a-readme-content",
		},
		{
			name: "should collect the last file if there is multiple readme",
			args: []args{
				{
					rel: "Readme.md",
					r:   strings.NewReader("a-readme-content"),
				},
				{
					rel: "readme",
					r:   strings.NewReader("another-readme-content"),
				},
			},
			want: "another-readme-content",
		},
		{
			name: "should collect the last readme even if the last is not a readme",
			args: []args{
				{
					rel: "Readme.md",
					r:   strings.NewReader("a-readme-content"),
				},
				{
					rel: "another",
					r:   strings.NewReader("something that is not a readme"),
				},
			},
			want: "a-readme-content",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := powerpacks.NewReadmeProcessor()

			for _, arg := range tt.args {
				assert.NoError(t, pr.Collect(context.TODO(), arg.rel, arg.r))
			}

			have, _ := io.ReadAll(pr.Readme)
			assert.Equal(t, tt.want, string(have))
		})
	}
}

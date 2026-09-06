package powerpacks_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/merlindorin/tk/pkg/powerpacks"
)

func testTaskfiles() map[string]string {
	return map[string]string{
		"git":      ".tk/git/Taskfile.yaml",
		"golangci": ".tk/golangci/Taskfile.yaml",
	}
}

func TestMergeTaskfile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "should create a taskfile when there is none",
			content: "",
			want: `version: '3'

includes:
    git: .tk/git/Taskfile.yaml #!tk
    golangci: .tk/golangci/Taskfile.yaml #!tk

dotenv:
  - .env
  - PROJECT
  - .env.default
`,
		},
		{
			name: "should keep everything tk does not own",
			content: `version: '3'

vars:
  APP: tk

includes:
    git: .tk/git/Taskfile.yaml #!tk
    stale: .tk/stale/Taskfile.yaml #!tk

tasks:
  build:
    cmds:
      - go build ./...
`,
			want: `version: '3'

vars:
  APP: tk

includes:
    git: .tk/git/Taskfile.yaml #!tk
    golangci: .tk/golangci/Taskfile.yaml #!tk

tasks:
  build:
    cmds:
      - go build ./...
`,
		},
		{
			name: "should adopt includes generated before the marker existed",
			content: `version: '3'

includes:
    git: .tk/git/Taskfile.yaml
    golangci: .tk/golangci/Taskfile.yaml

dotenv:
  - .env
`,
			want: `version: '3'

includes:
    git: .tk/git/Taskfile.yaml #!tk
    golangci: .tk/golangci/Taskfile.yaml #!tk

dotenv:
  - .env
`,
		},
		{
			name: "should preserve the includes the user wrote",
			content: `version: '3'

includes:
  # our own tasks
  docker: ./build/docker/Taskfile.yaml
  golangci: .tk/golangci/Taskfile.yaml
  deploy:
    taskfile: ./deploy/Taskfile.yaml
    optional: true
`,
			want: `version: '3'

includes:
  # our own tasks
  docker: ./build/docker/Taskfile.yaml
  git: .tk/git/Taskfile.yaml #!tk
  golangci: .tk/golangci/Taskfile.yaml #!tk
  deploy:
    taskfile: ./deploy/Taskfile.yaml
    optional: true
`,
		},
		{
			name: "should add an includes block when there is none",
			content: `version: '3'

tasks:
  build:
    cmds:
      - go build ./...
`,
			want: `version: '3'

tasks:
  build:
    cmds:
      - go build ./...

includes:
    git: .tk/git/Taskfile.yaml #!tk
    golangci: .tk/golangci/Taskfile.yaml #!tk
`,
		},
		{
			name: "should append to an empty includes block",
			content: `version: '3'

includes:

dotenv:
  - .env
`,
			want: `version: '3'

includes:
    git: .tk/git/Taskfile.yaml #!tk
    golangci: .tk/golangci/Taskfile.yaml #!tk

dotenv:
  - .env
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := powerpacks.MergeTaskfile(tt.content, testTaskfiles())
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMergeTaskfile_isIdempotent(t *testing.T) {
	first, err := powerpacks.MergeTaskfile("", testTaskfiles())
	assert.NoError(t, err)

	second, err := powerpacks.MergeTaskfile(first, testTaskfiles())
	assert.NoError(t, err)
	assert.Equal(t, first, second)
}

func TestMergeTaskfile_rejectsInlineIncludes(t *testing.T) {
	_, err := powerpacks.MergeTaskfile("version: '3'\n\nincludes: {}\n", testTaskfiles())
	assert.IsError(t, err, powerpacks.ErrInlineIncludes)
}

func TestMergeEnvrc(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "should not create an envrc",
			content: "",
			want:    "",
		},
		{
			name:    "should remove the export tk used to write",
			content: "export TASK_X_REMOTE_TASKFILES=1 #!tk\n",
			want:    "",
		},
		{
			name:    "should remove it from a file written before the marker existed",
			content: "export TASK_X_REMOTE_TASKFILES=1\n",
			want:    "",
		},
		{
			name:    "should keep the exports the user wrote",
			content: "use flake\nexport TASK_X_REMOTE_TASKFILES=1 #!tk\nexport FOO=bar\n",
			want:    "use flake\nexport FOO=bar\n",
		},
		{
			name:    "should leave a file it owns nothing in alone",
			content: "use flake\nexport FOO=bar\n",
			want:    "use flake\nexport FOO=bar\n",
		},
		{
			name:    "should keep windows line endings",
			content: "use flake\r\nexport TASK_X_REMOTE_TASKFILES=1 #!tk\r\n",
			want:    "use flake\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, powerpacks.MergeEnvrc(tt.content))
		})
	}
}

func TestMergeEnvrc_neverWritesTheReleasedExperiment(t *testing.T) {
	// Task released the REMOTE_TASKFILES experiment and warns about the variable on every
	// invocation, so tk must never put it back into a project.
	for _, content := range []string{"", "use flake\n", "export TASK_X_REMOTE_TASKFILES=1\n"} {
		assert.NotContains(t, powerpacks.MergeEnvrc(content), "TASK_X_REMOTE_TASKFILES")
	}
}

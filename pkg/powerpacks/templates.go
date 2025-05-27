package powerpacks

const (
	envrcTemplate    = `export TASK_X_REMOTE_TASKFILES=1`
	taskfileTemplate = `version: '3'

includes:
    {{ range $name, $filename := . }}
        {{- $name }}: {{ $filename }}
    {{end}}

dotenv:
  - .env
  - PROJECT
  - .env.default`
)

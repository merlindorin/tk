package powerpacks

const (
	envrcTemplate = `export TASK_X_REMOTE_TASKFILES=1`

	aquaTemplate = `---
# aqua - Declarative CLI Version Manager
# https://aquaproj.github.io/
# checksum:
#   enabled: true
#   require_checksum: true
#   supported_envs:
#   - all
registries:
- type: standard
  ref: v4.219.0 # renovate: depName=aquaproj/aqua-registry

packages:
  - name: go-task/task@v3.38.0
  - name: jqlang/jq@jq-1.7.1
  {{ range $name, $filename := . -}}
  - import: {{ $filename }}
  {{ end }}
`
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

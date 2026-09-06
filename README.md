# tk

[![Build Status](https://github.com/merlindorin/tk/actions/workflows/golangci.yml/badge.svg)](https://github.com/merlindorin/tk/actions/workflows/golangci.yml)
[![Release Status](https://github.com/merlindorin/tk/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/merlindorin/tk/actions/workflows/goreleaser.yml)

> A CLI tool that bootstraps your projects with pre-configured [Taskfile](https://taskfile.dev) powerpacks. Stop copying boilerplate configs between projects—install what you need in seconds.

## Table of Content

* [Features](#features)
* [Prerequisites](#prerequisites)
* [Installation](#installation)
  * [Binaries via GitHub Releases](#binaries-via-github-releases)
  * [Binary Installation using Script](#binary-installation-using-script)
  * [Docker Image](#docker-image)
* [Usage](#usage)
  * [Owned lines](#owned-lines)
* [Development](#development)
  * [Repository Structure](#repository-structure)
  * [Development with Taskfile](#development-with-taskfile)
* [Contributing](#contributing)
* [License](#license)

## Features

- **Powerpacks**: Curated bundles of Taskfile configurations for common tools (linters, formatters, CI workflows)
- **Self-updating**: pin `tk` in `go.mod` and let Dependabot bump it, with a workflow that regenerates and opens a PR
- **Zero config**: Sensible defaults that work out of the box
- **Composable**: Mix and match powerpacks to fit your stack, add and remove them at any time
- **Consistent**: Same tooling setup across all your projects
- **Non-destructive**: Only the lines `tk` generated are rewritten; your own tasks and includes stay untouched
- **Predictable**: `tk status` and `--dry-run` show every change before it happens

## Prerequisites

To make the most out of this project, ensure the following tools are installed:

- [Git](https://git-scm.com): Essential for version control and managing codebase changes.
- [Task](https://taskfile.dev/): A task runner facilitating automated workflows and tasks (v3.38.0 or later).
- [jq](https://jqlang.github.io/jq/): A lightweight and flexible command-line JSON processor (v1.7.1 or later).

## Installation

### Binaries via GitHub Releases

1. Visit the [GitHub Releases page](https://github.com/merlindorin/tk/releases) of this repository.
2. Download the appropriate binary for your operating system.
3. Make the downloaded file executable:
   ```bash
   chmod +x tk
   ```
4. Move it to a location within your PATH, such as `/usr/local/bin`, for easy access.

### Binary Installation using Script

1. Install `tk` using the installation script:
   ```bash
   # binary will be installed in $(go env GOPATH)/bin/tk
   curl -sSfL https://raw.githubusercontent.com/merlindorin/tk/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
   tk --version
   ```

2. Note for Windows users: You can run the above commands with Git Bash, which comes with Git for Windows.

### Docker Image

1. Pull the Docker image directly from the GitHub repository:
   ```bash
   docker pull ghcr.io/merlindorin/tk:latest
   ```
2. Run the `tk` CLI using Docker:
   ```bash
   docker run ghcr.io/merlindorin/tk:latest [command]
   ```

## Usage

Use the `tk` CLI to install powerpacks, which include Taskfiles. These files are tailored for efficient
management of tools and environments, easing development workflows.

```bash
tk init                          # install every powerpack in the current project
tk init --include go,golangci    # ... or only these ones
tk update                        # refresh the project from the powerpacks of the binary
tk status                        # what would `tk update` change?
tk list                          # available powerpacks, their task prefix and their state
tk list --tasks                  # ... with the tasks each one provides
tk add golangci                  # install one more powerpack
tk remove trufflehog             # remove one
```

`--include` is an allow list: leave it out and every powerpack is installed, name some and only those are.
`tk list` shows the prefix each powerpack's tasks live under:

```console
$ tk list
NAME        PREFIX       STATE      DESCRIPTION
git         git:         installed  Taskfile Git Pre-Commit Hook
go          go:          installed  Golang Taskfile
golangci    golangci:    installed  Golang CI Taskfile
dependabot  dependabot:  available  Dependabot Taskfile

$ tk list --tasks
NAME        PREFIX       STATE      TASKS
go          go:          installed  go:build go:ci go:cover go:default go:pre-commit go:test go:tidy
```

Every command accepts `--target <dir>` to work on another project, and the commands that write accept `--dry-run`
to report the changes without touching anything:

```console
$ tk status
tk v0.3.0 (project written with v0.2.0)
8 powerpacks installed, 1 excluded

  update  .tk/golangci/Taskfile.yaml
  update  Taskfile.yaml

2 changes to apply, 15 unchanged (dry run, nothing written)
run `tk update` to apply
```

`.tk.yaml` records the options, the selected powerpacks, the tk version that wrote the project and a checksum per
installed powerpack, so `tk status` and `tk list` can tell you what is outdated. A project written by an older tk
with the previous `excludes:` deny list is converted to `includes:` on the next write, keeping the same powerpacks. `tk status --check` exits non-zero
when the project is out of date, which makes it usable as a CI gate.

`tk` merges its includes into the root Taskfile the project already uses — `Taskfile.yml`, `Taskfile.yaml` or one of
the `dist` variants, in the order Task itself resolves them — and only creates `Taskfile.yaml` when there is none.

### Owned lines

`tk init` and `tk update` never rewrite your whole `Taskfile.yaml`: they only touch the lines they generated,
which are marked with a trailing `#!tk` comment.

```yaml
version: '3'

includes:
  docker: ./build/docker/Taskfile.yaml       # yours, left untouched
  golangci: .tk/golangci/Taskfile.yaml #!tk  # tk's, refreshed on update

tasks:
  build:                                     # yours, left untouched
    cmds:
      - go build ./...
```

Everything else — your own includes, variables, tasks, comments and formatting — is preserved. Removing a powerpack
(via `tk remove`) drops its marked include and nothing more. Includes generated by older versions of `tk`
are adopted and marked on the next `tk update`. The `.tk/` directory stays fully owned by `tk` and is regenerated on
every run.

`tk` no longer writes to `.envrc`. It used to export `TASK_X_REMOTE_TASKFILES=1` to turn on a Task experiment;
Task released the experiment and now prints a warning about the variable on every invocation, so `tk update`
removes that line instead — and deletes the file when it held nothing else. Pass `--disable-envrc` to leave the
file untouched. If your shell still exports the variable, direnv is holding the old value: reload it or open a
new shell.

## Development

To develop and test features for the `tk` CLI:

1. Clone this repository:
   ```bash
   git clone https://github.com/merlindorin/tk.git
   cd tk
   ```

2. Install necessary dependencies and tools.
3. Use a feature branch for your development:
   ```bash
   git checkout -b feature/my-new-feature
   ```
4. Develop and test your changes locally.
5. Commit your changes with descriptive messages.

### Development with Taskfile

Leverage the Taskfile in this repository to automate common development tasks:

```bash
task go:build       # build the CLI into ./bin
task go:test        # run the test suite
task go:cover       # run the tests and print the coverage report
task golangci:lint  # lint the codebase
task                # run every `*:default` task
```

Run `task --list-all` for the full list, and review the `Taskfile.yaml` in the root directory to understand how the
powerpacks are wired together.

After changing anything under `powerpacks/`, rebuild the CLI and run `tk status` followed by `tk update` in this
repository so `.tk/` reflects the change.

## Contributing

Interested in contributing?

- Fork the repository.
- Create a branch for your feature: `git checkout -b feature/your-feature`.
- Commit your changes: `git commit -am 'Add a feature'`.
- Push to your branch: `git push origin feature/your-feature`.
- Open a pull request for review.

## License

Licensed under the MIT License. See [LICENSE.md](./LICENSE.md) for further information.

# Golang Taskfile

> The provided Taskfile is intended to be a reusable component that houses the everyday Go commands of a repository:
> building, testing, coverage and the GitHub Action running them.

<!-- TOC -->

* [Golang Taskfile](#golang-taskfile)
  * [Summary](#summary)
  * [Tasks](#tasks)
  * [Variables](#variables)
  * [Usage](#usage)

<!-- TOC -->

## Summary

This powerpack wraps the `go` toolchain itself. It does not install anything: it only standardises how a project is
built and tested, so every repository answers to the same `task go:test` regardless of its layout.

## Tasks

| TASK      | DESCRIPTION                                                     |
|-----------|-----------------------------------------------------------------|
| `default` | Run the test suite then generate the GitHub Action              |
| `build`   | Build every main package into the build directory               |
| `test`    | Run the test suite                                              |
| `cover`   | Run the test suite with coverage and print the per-function report |
| `tidy`    | Tidy the go module                                              |
| `ci`      | Generate the GitHub Action running build and tests              |
| `pre-commit` | Run the test suite before a commit                           |

## Variables

| VARIABLE            | DESCRIPTION                | DEFAULT           |
|---------------------|----------------------------|-------------------|
| `GO_BIN`            | go binary                  | `go`              |
| `SOURCES`           | go sources of the project  | `./...`           |
| `BUILD_DIR`         | where binaries are written | `bin`             |
| `TARGET`            | packages to build          | `./cmd/...`       |
| `TEST_ARGS`         | go test arguments          | `-race`           |
| `COVERAGE_FILENAME` | coverage profile           | `coverage.out`    |

## Usage

```bash
task go:test                    # go test -race ./...
task go:test -- -run TestFoo    # run a single test
task go:cover                   # coverage report
task go:build                   # build every command into ./bin
TARGET=./cmd/tk task go:build   # build a single command
```

The `ci` task writes `.github/workflows/go.yml`. The file is only generated when it does not exist yet, so local
changes are never overwritten; delete it to regenerate a fresh one.

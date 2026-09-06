# tk Taskfile

> The provided Taskfile is intended to be a reusable component that keeps `tk` itself up to date in a repository,
> and keeps the files it generates in sync with the version it is pinned to.

<!-- TOC -->

* [tk Taskfile](#tk-taskfile)
  * [Summary](#summary)
  * [How the automation works](#how-the-automation-works)
  * [Tasks](#tasks)
  * [Variables](#variables)
  * [Usage](#usage)
  * [Without Dependabot](#without-dependabot)

<!-- TOC -->

## Summary

Updating `tk` in a repository is two separate problems:

1. **Getting a newer `tk`** — solved by pinning it in the `tool` block of `go.mod`, where Dependabot sees it.
2. **Regenerating the files it produces** — a newer `tk` changes nothing until `tk update` runs.

This powerpack does both: `install` pins the binary, and the workflow it ships regenerates `.tk/` and opens a pull
request when it drifts.

It also requires the `go` powerpack, so installing it pulls that in: the workflow runs `go tool tk`, which needs
the module wiring `go` provides.

## How the automation works

```
  Dependabot  ──weekly──▶  bumps tk in go.mod            (gomod ecosystem, see the dependabot powerpack)
  tk workflow ──weekly──▶  go tool tk update ──▶ pull request when .tk/ changed
  tk workflow ──on PR──▶   go tool tk status --check ──▶ fails when .tk/ is stale
```

The two are independent: the scheduled job also catches a powerpack that changed without a version bump, and the
pull request check makes a stale `.tk/` visible before it is merged.

Pinning with `go get -tool` rather than `go install` is what makes the first line work: tool directives live in
`go.mod`, so Dependabot treats `tk` like any other Go dependency. It also pins the exact `tk` everyone on the team
and CI runs, which is what makes the drift check meaningful.

## Tasks

| TASK         | DESCRIPTION                                                       |
|--------------|-------------------------------------------------------------------|
| `default`    | Fail when the powerpacks are out of date                          |
| `install`    | Pin `tk` in the `tool` block of `go.mod`                          |
| `uninstall`  | Remove the tool directive                                         |
| `run`        | Run the pinned `tk` with arbitrary arguments                      |
| `update`     | Regenerate the powerpacks (`tk update`)                           |
| `status`     | Fail when the powerpacks are out of date (`tk status --check`)    |
| `pre-commit` | Fail the commit when the powerpacks are out of date               |

## Variables

| VARIABLE   | DESCRIPTION                          | DEFAULT                         |
|------------|--------------------------------------|---------------------------------|
| `GO_BIN`   | go binary                            | `go`                            |
| `PACKAGE`  | tk package                           | `github.com/merlindorin/tk/cmd/tk` |
| `VERSION`  | version to pin                       | `latest`                        |

## Usage

```bash
task tk:install                 # pin tk in go.mod — Dependabot now tracks it
task tk:status                  # is .tk/ in sync with the pinned tk?
task tk:update                  # regenerate it
task tk:run -- list             # any other tk command
```

`.github/workflows/tk.yml` is written by `tk` itself, not by a task: the powerpack manifest declares it with the
`sync` strategy, so a newer `tk` can fix its own automation on the next update. Editing it is pointless — the file
is put back. Installing the powerpack writes it, removing the powerpack deletes it.

The workflow needs nothing but the default `GITHUB_TOKEN`. If your repository requires pull requests to be opened
by a real account, or has "Allow GitHub Actions to create pull requests" disabled, you need a personal access token
instead — and since `tk` owns the file, editing it in place would be undone on the next update. Take the file over
instead: edit it, then run `tk remove tk`. An owned file the project has edited is reported and left behind rather
than deleted, so it becomes yours to maintain.

## Without Dependabot

Nothing here requires Dependabot. The scheduled job alone keeps the generated files in sync with the pinned `tk`,
and you can bump the pin by hand:

```bash
task tk:install VERSION=latest && task tk:update
```

If you would rather not add `tk` to `go.mod` at all, install it globally and drop this powerpack:

```bash
go install github.com/merlindorin/tk/cmd/tk@latest && tk update
```

That version is invisible to Dependabot, so keeping it current is on you.

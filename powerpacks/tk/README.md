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

This powerpack does both: `install` pins the binary, and `ci` generates a GitHub Action that regenerates `.tk/`
and opens a pull request when it drifts.

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
| `default`    | Generate the GitHub Action                                        |
| `install`    | Pin `tk` in the `tool` block of `go.mod`                          |
| `uninstall`  | Remove the tool directive                                         |
| `run`        | Run the pinned `tk` with arbitrary arguments                      |
| `update`     | Regenerate the powerpacks (`tk update`)                           |
| `status`     | Fail when the powerpacks are out of date (`tk status --check`)    |
| `ci`         | Generate the GitHub Action running both of the above              |
| `pre-commit` | Fail the commit when the powerpacks are out of date               |

## Variables

| VARIABLE   | DESCRIPTION                          | DEFAULT                         |
|------------|--------------------------------------|---------------------------------|
| `GO_BIN`   | go binary                            | `go`                            |
| `PACKAGE`  | tk package                           | `github.com/merlindorin/tk/cmd/tk` |
| `VERSION`  | version to pin                       | `latest`                        |
| `SCHEDULE` | cron expression of the update job    | `0 6 * * 1` (Monday morning)    |
| `BRANCH`   | branch holding the generated updates | `tk/update`                     |

## Usage

```bash
task tk:install                 # pin tk in go.mod — Dependabot now tracks it
task tk:ci                      # write .github/workflows/tk.yml
task tk:status                  # is .tk/ in sync with the pinned tk?
task tk:update                  # regenerate it
task tk:run -- list             # any other tk command
```

The generated workflow needs nothing but the default `GITHUB_TOKEN`. If your repository requires pull requests to
be opened by a real account, or has "Allow GitHub Actions to create pull requests" disabled, replace `github.token`
with a personal access token in `.github/workflows/tk.yml`.

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

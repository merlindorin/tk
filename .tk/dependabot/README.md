# Dependabot Taskfile

> The provided Taskfile is intended to be a reusable component that generates the Dependabot configuration of a
> repository, keeping Go modules, pinned tools and GitHub Actions up to date.

<!-- TOC -->

* [Dependabot Taskfile](#dependabot-taskfile)
  * [Summary](#summary)
  * [Tasks](#tasks)
  * [Variables](#variables)
  * [What gets updated](#what-gets-updated)
  * [Usage](#usage)

<!-- TOC -->

## Summary

Dependabot is configured by a single file, `.github/dependabot.yml`, and GitHub runs it for you. This powerpack
writes that file with sane defaults: weekly checks, one grouped pull request per ecosystem, and conventional commit
messages.

Updates are **grouped**: without grouping, a Go module with a large indirect dependency list produces dozens of pull
requests a week. With grouping you get one `chore(deps)` pull request per ecosystem.

## Tasks

| TASK         | DESCRIPTION                          |
|--------------|--------------------------------------|
| `default`    | Generate the configuration           |
| `config`     | Generate `.github/dependabot.yml`    |
| `pre-commit` | Generate the configuration if missing |

## Variables

| VARIABLE          | DESCRIPTION                            | DEFAULT          |
|-------------------|----------------------------------------|------------------|
| `CONFIG_FILENAME` | configuration filename in `.github`    | `dependabot.yml` |
| `INTERVAL`        | how often Dependabot looks for updates | `weekly`         |
| `COMMIT_PREFIX`   | conventional commit prefix             | `chore`          |

## What gets updated

| ECOSYSTEM        | WHAT IT COVERS                                                                     |
|------------------|------------------------------------------------------------------------------------|
| `gomod`          | `require` blocks **and the `tool` block** of `go.mod` — golangci-lint, goreleaser, `tk` itself |
| `github-actions` | the actions used by the workflows in `.github/workflows`                            |
| `docker`         | base images in a `Dockerfile` (commented out by default)                            |

The `gomod` entry is what keeps a pinned tool current: anything installed with `go get -tool` lives in `go.mod`, so
Dependabot sees it like any other dependency. See the `tk` powerpack to pin `tk` itself that way.

## Usage

```bash
task dependabot:config              # write .github/dependabot.yml
INTERVAL=daily task dependabot:config
```

The file is generated only when it does not exist yet, so local changes are never overwritten. Delete it to
regenerate a fresh one, or edit it directly to ignore a dependency:

```yaml
    ignore:
      - dependency-name: "github.com/example/pinned"
        versions: ["2.x"]
```

# Monorepository for Reusable Taskfiles

This repository contains a collection of reusable Taskfiles for various development workflows. Each Taskfile is designed
to encapsulate best practices and tools for specific tasks, making it easy to integrate consistent and reliable commands
across multiple projects.

<!-- TOC -->

* [Monorepository for Reusable Taskfiles](#monorepository-for-reusable-taskfiles)
    * [Summary](#summary)
    * [Prerequisites](#prerequisites)
    * [Repository Structure](#repository-structure)
    * [Available Taskfiles](#available-taskfiles)
    * [Usage](#usage)
    * [Contributing](#contributing)
    * [License](#license)

<!-- TOC -->

## Summary

This monorepository aims to simplify project setup and maintenance by providing a collection of pre-configured
Taskfiles. These Taskfiles cover a variety of use-cases, such as linting, testing, building, and more. By including
these Taskfiles in your projects, you can ensure consistency and save time by not having to reinvent the wheel for each
new project.

## Prerequisites

To use the Taskfiles in this repository, you need to have the following tools installed:

* [Git](https://git-scm.com): A Version manager
* [Task](https://taskfile.dev/): A task runner / simpler Make alternative written in Go.
* [Aqua](https://aquaproj.github.io): A CLI manager / ensure CLIs are runnable and up to date.

Before you proceed, make sure you have these tools installed.

## Repository Structure

The repository is organized as follows:

```
monorepo/
├── README.md
├── taskfiles/
│   ├── golangci
│   ├── git
│   ├── default
│   ├── goreleaser
│   ├── trufflehog
│   └── etc...
└── examples/
    ├── example1/
    │   └── Taskfile.yml
    └── example2/
        └── Taskfile.yml
```

* `taskfiles/`: Contains individual folder with Taskfiles for different purposes.
* `examples/`: Contains example projects showing how to include and use the Taskfiles.

## Available Taskfiles

Here are some of the Taskfiles available in this repository:

* **Golang CI (`taskfiles/golangci.yaml`)**: Linting, fixing, and boilerplate setup for Golang projects
  using `golangci-lint`.
* **Python Lint (`taskfiles/pythonlint.yaml`)**: Linting for Python projects using tools like `flake8`, `pylint`, etc.
* **Docker (`taskfiles/docker.yaml`)**: Docker-related tasks such as building, pushing, and running containers.

Each Taskfile comes with its own set of tasks and configurations. Refer to the respective Taskfile for detailed
documentation.

## Usage

To use a Taskfile from this repository, include it in your project's `Taskfile.yml`. For example, to include the Golang
CI Taskfile, add the following to your `Taskfile.yml`:

```yaml
version: '3'

includes:
  golangci: https://raw.githubusercontent.com/merlindorin/taskfiles/main/taskfiles/golangci.yaml
```

Then, you can run the tasks defined in the included Taskfile as follows:

```sh
task golangci:lint
task golangci:fix
```

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository.
2. Create a new branch with a descriptive name.
3. Add your changes.
4. Create a pull request.

Make sure to follow the contribution guidelines and ensure your Taskfile is well-documented.

## License

This repository is licensed under the MIT License. See the [LICENSE](LICENSE) file for more information.

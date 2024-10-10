# tk: Simplify Tool Installation for Projects

[![Build Status](https://github.com/merlindorin/tk/actions/workflows/golangci.yml/badge.svg)](https://github.com/merlindorin/tk/actions/workflows/golangci.yml)
[![Test Status](https://github.com/merlindorin/tk/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/merlindorin/tk/actions/workflows/goreleaser.yml)
[![Test Status](https://github.com/merlindorin/tk/actions/workflows/trufflehog.yml/badge.svg)](https://github.com/merlindorin/tk/actions/workflows/trufflehog.yml)

Welcome to `tk`, a command-line interface (CLI) designed to streamline the installation of tools for your projects.
Managed by [merlindorin](https://github.com/merlindorin), this project offers simplicity and efficiency by using
powerpacks that bundle necessary configurations for Taskfile and Aquafiles.

## Table of Contents

- [Summary](#summary)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
    - [Binaries via GitHub Releases](#binaries-via-github-releases)
    - [Binary Installation using Script](#binary-installation-using-script)
    - [Docker Image](#docker-image)
- [Usage](#usage)
- [Development](#development)
    - [Repository Structure](#repository-structure)
    - [Development with Taskfile](#development-with-taskfile)
    - [Installing Tools with Aqua](#installing-tools-with-aqua)
- [Contributing](#contributing)
- [License](#license)

## Summary

`tk` simplifies your project setup by providing powerpacks that ensure consistent configuration across your development
environment. This ensures you spend less time managing setups and more time focusing on development.

## Prerequisites

To make the most out of this project, ensure the following tools are installed:

- [Git](https://git-scm.com): Essential for version control and managing codebase changes.
- [Task](https://taskfile.dev/): A task runner facilitating automated workflows and tasks.
- [Aqua](https://aquaproj.github.io): Ensures your CLI tools are up-to-date.

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

Use the `tk` CLI to install powerpacks, which include Taskfiles and Aquafiles. These files are tailored for efficient
management of tools and environments, easing development workflows.

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

### Repository Structure

- `cmd/tk`: Contains the main source code for the `tk` CLI.
- `powerpacks`: Configuration folders that include Taskfiles and Aquafiles.
- `Taskfile.yaml`: Main Taskfile for automating development tasks.
- `aqua.yaml`: Configuration file specifying CLI tools for Aqua.

### Development with Taskfile

Leverage the Taskfile in this repository to automate common development tasks. Typical tasks available include building
the CLI, running tests, and linting the codebase. Review the `Taskfile.yaml` available in the root directory to
understand the tasks and their configuration.

### Installing Tools with Aqua

1. **Install Aqua**: Follow the instructions on the [Aqua installation guide](https://aquaproj.github.io/docs/install)
   to set up Aqua CLI.

2. **Configure Aqua**: Use the `aqua.yaml` file present in the repository to define the CLI tools and their versions for
   your development setup.

3. **Install Tools**: Run the following command to install all necessary tools listed in the `aqua.yaml`:
   ```bash
   aqua i
   ```

   Aqua will ensure all specified tools are installed in your environment, leveraging its centralized configuration for
   consistency across systems.

## Contributing

Interested in contributing?

- Fork the repository.
- Create a branch for your feature: `git checkout -b feature/your-feature`.
- Commit your changes: `git commit -am 'Add a feature'`.
- Push to your branch: `git push origin feature/your-feature`.
- Open a pull request for review.

## License

Licensed under the MIT License. See [LICENSE.md](./LICENSE.md) for further information.

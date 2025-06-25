# Contributing to CrowNet

First off, thank you for considering contributing to CrowNet! We welcome any contributions that can help improve the project, whether it's reporting a bug, proposing a new feature, improving documentation, or writing code.

This document provides guidelines to help you contribute effectively.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Enhancements](#suggesting-enhancements)
  - [Your First Code Contribution](#your-first-code-contribution)
  - [Pull Requests](#pull-requests)
- [Development Setup](#development-setup)
- [Style Guides](#style-guides)
  - [Go Code](#go-code)
  - [Commit Messages](#commit-messages)
- [Testing](#testing)
- [Building the Project](#building-the-project)

## Code of Conduct

This project and everyone participating in it is governed by a [Code of Conduct](CODE_OF_CONDUCT.md) (Note: CODE_OF_CONDUCT.md needs to be created, TBD). By participating, you are expected to uphold this code. Please report unacceptable behavior.

## How Can I Contribute?

### Reporting Bugs

If you find a bug, please ensure the bug was not already reported by searching on GitHub under [Issues](https://github.com/your-repo/crownet/issues). <!-- TODO: Replace with actual repo link -->

If you're unable to find an open issue addressing the problem, [open a new one](https://github.com/your-repo/crownet/issues/new). Be sure to include a **title and clear description**, as much relevant information as possible, and a **code sample or an executable test case** demonstrating the expected behavior that is not occurring.

<!-- TODO: When CHORE-006 (Issue Templates) is done, link to the bug report template here. -->

### Suggesting Enhancements

If you have an idea for an enhancement or a new feature, please outline your proposal in an issue on GitHub. This allows for discussion and refinement before any significant development work begins.

<!-- TODO: When CHORE-006 (Issue Templates) is done, link to the feature request template here. -->

### Your First Code Contribution

Unsure where to begin contributing to CrowNet? You can start by looking through `good first issue` and `help wanted` issues:

- [Good first issues](https://github.com/your-repo/crownet/labels/good%20first%20issue) - issues which should only require a few lines of code, and a test or two. <!-- TODO: Replace link -->
- [Help wanted issues](https://github.com/your-repo/crownet/labels/help%20wanted) - issues which should be a bit more involved than `good first issue` issues. <!-- TODO: Replace link -->

### Pull Requests

1.  Fork the repository and create your branch from `main`.
2.  If you've added code that should be tested, add tests.
3.  If you've changed APIs, update the documentation.
4.  Ensure the test suite passes (see [Testing](#testing)).
5.  Make sure your code lints (see [Style Guides](#style-guides)).
6.  Issue that pull request!

## Development Setup

Please refer to the [Environment Setup Guide](docs/03_guias/guia_configuracao_ambiente.md) for instructions on how to set up your development environment. Key requirements include:
- Go (version specified in `go.mod`)
- Any other tools mentioned in the setup guide.

## Style Guides

### Go Code

All Go code should adhere to the guidelines outlined in our [Go Code Style Guide](docs/03_guias/guia_estilo_codigo.md).
Before submitting code, please run `go fmt`. You can use the Makefile to run the linter:
`make lint`
This command uses `golangci-lint` (which you may need to install: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`).

*(Note: The `run_in_bash_session` tool required for running linters via make is currently experiencing issues. You might need to run linters manually if `make lint` fails due to tool errors.)*

### Commit Messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for our commit messages. This allows for easier history tracking and automated changelog generation.

A commit message should be structured as follows:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Example:
`feat(network): add support for neuron refractory periods`
`fix(config): correct default value for learning rate`
`docs(readme): update project description`

## Testing

All contributions that include new features or bug fixes should be accompanied by unit tests.
You can run the test suite using the Makefile:
`make test`
This command will run all tests verbosely (`go test ./... -v`). Ensure all tests pass before submitting a pull request.

*(Note: The `run_in_bash_session` tool required for running `make test` or `go test` is currently experiencing issues. Test execution might be temporarily hindered.)*

## Building the Project

To build the project executable (`crownet`), you can use the Makefile:
`make build`
This will compile `main.go` and place the output binary named `crownet` in the project root.

Alternatively, you can use the standard Go command:
`go build -o crownet main.go`

*(Note: The `run_in_bash_session` tool required for building via `make build` or `go build` is currently experiencing issues.)*

---

We look forward to your contributions!

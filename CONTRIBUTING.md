# Contributing to Opservant Spark

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/opservant-spark.git`
3. Create a branch: `git checkout -b feat/your-feature`
4. Make your changes
5. Open a pull request against `main`

## Development Setup

**Requirements:** Go 1.26+

```bash
git clone https://github.com/s4e-io/opservant-spark.git
cd opservant-spark
go mod download
go build -o spark ./cmd/opservant-spark
```

## Before Submitting

```bash
go vet ./...
go test ./...
```

Make sure both pass with no errors.

## Pull Request Guidelines

- Keep PRs focused — one feature or fix per PR
- Add or update example playbooks in `examples/` if your change affects playbook behavior
- Write a clear PR description: what changed and why

## Reporting Bugs

Open a [GitHub Issue](https://github.com/s4e-io/opservant-spark/issues) with:
- OS and architecture
- Go version (`go version`)
- Steps to reproduce
- Expected vs actual behavior

## Playbook Contributions

New playbooks belong in the [opservant-playbooks](https://github.com/s4e-io/opservant-playbooks) repository, not here.

## Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/). Every commit must follow:

```
<type>(<scope>): <description>
```

Common types:

| Type | When to use |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `chore` | Build, deps, tooling |
| `refactor` | Code change, no feature/fix |
| `test` | Adding or updating tests |

Examples:
```
feat: add --timeout flag to playbook command
fix: handle missing config.yaml gracefully
docs: update README installation steps
chore: upgrade viper to v1.17
```

Breaking changes: add `!` after type or `BREAKING CHANGE:` in footer:
```
feat!: rename --playbook-dir to --dir
```

Commit messages drive the automated CHANGELOG and version bumps — `feat` bumps minor, `fix` bumps patch, breaking change bumps major.

## Code Style

Follow standard Go conventions. Run `go vet` before opening a PR.

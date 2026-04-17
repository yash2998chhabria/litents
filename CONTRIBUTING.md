# Contributing to Litents

Thank you for your interest in contributing to Litents.

## Getting started

- Install Go 1.22+
- From the repo root run:

```bash
go test ./...
```

- For local development, run:

```bash
go run ./cmd/litents --help
```

## Reporting issues

Please include:

- Command used
- OS and architecture
- Output from `litents doctor`
- Steps to reproduce

## Pull requests

- Keep changes focused and small.
- Add or update tests for behavioral changes.
- Update docs when commands, flags, or outputs change.
- Ensure CI is green before requesting review.

## Code style

- Keep the project lightweight and avoid adding heavy dependencies.
- Prefer small, testable helper functions.
- Preserve plain-file storage layout and tmux/git command behavior.

# Project: km_build_upload

## Overview

`km_build_upload` is a custom GitHub Action that replaces broken third-party actions for automated versioning and Docker image publishing. It handles semantic version bumping, version file management, and git tagging — then exposes outputs for downstream Docker build/push steps using Docker's official actions.

## Motivation

Third-party GitHub Actions (like `anothrNick/github-tag-action` and `dolittle/write-version-to-file-action`) broke and are no longer maintained. This project replaces them with a single, self-contained composite action written in Go.

## Architecture

### Composite Action + Go

The action uses GitHub's composite action mechanism to run a Go program directly on the runner. This avoids Docker-in-Docker complexity while giving us the robustness of a compiled language.

- **`action.yml`**: Defines the composite action, runs `go run main.go`, and maps outputs.
- **`main.go`**: All logic in a single file — version parsing, file I/O, git operations, and output writing.
- **No external Go dependencies**: Uses only the Go standard library.

### Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Action type | Composite | Avoids Docker-in-Docker; runs Go directly on the runner |
| Language | Go (stdlib only) | Team preference; robust error handling; no dependency management |
| Version source of truth | `version` file | Simpler than git tags; always visible in repo; no `fetch-depth: 0` needed |
| Version format | `1.2.3` (no `v` prefix) | Team preference |
| Docker steps | Delegated to Docker's official actions | Well-maintained; follows "do one thing well" philosophy |
| Error handling | Descriptive stderr messages with step context and hints | Focus on actionable error content |
| Git identity | `github-actions[bot]` | Standard GitHub bot identity |

### Flow

```
version file (read) → bump patch → version file (write) → git add → git commit → git tag → git push → write outputs
```

Downstream steps (in the caller workflow) use the outputs for Docker login, build, and push.

## Testing

- **Unit tests**: `go test ./...` — covers version parsing, bumping, file I/O, and repo name extraction.
- **Real-world testing**: Deploy the action to a test repo and trigger workflows.

## Files

| File | Purpose |
|---|---|
| `action.yml` | Composite action definition |
| `main.go` | Core Go logic |
| `main_test.go` | Unit tests |
| `go.mod` | Go module file |
| `examples/workflow.yml` | Ready-to-copy caller workflow |
| `implementations/high_level.impl` | Original spec + 4 rounds of review |
| `implementations/docker-image.yml` | Reference workflow being replaced |

## Future Enhancements

- `#Minor` / `#Major` commit message scanning for non-patch bumps
- Concurrent merge race condition handling (GitHub concurrency groups or retry logic)
- Configurable Dockerfile path and build context
- Configurable git identity

# km_build_upload

A custom, self-contained GitHub Action that bumps semantic versions, manages a version file, creates git tags, and exposes outputs for Docker image publishing — with zero third-party action dependencies.

## Features

- **Automatic Patch Versioning**: Reads a `version` file, bumps the patch number, and writes it back.
- **Git Tagging**: Commits the version bump and creates a matching git tag.
- **Action Outputs**: Exposes `new_version` and `image_name` for use in downstream Docker build/push steps.
- **Zero Third-Party Dependencies**: Only uses Go stdlib. Pairs with GitHub's official and Docker's official actions.
- **File as Source of Truth**: The `version` file in your repo root is the single source of truth — no need for deep git history fetches.

## Quick Start

### 1. Add the version file

Create a `version` file in your repo root (or let the action create it on first run):

```
0.0.0
```

### 2. Add the workflow

Copy `examples/workflow.yml` to your repo at `.github/workflows/docker-publish.yml`:

```yaml
name: Build and Publish Docker Image

on:
  push:
    branches: [ "main" ]

permissions:
  contents: write

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Bump version and tag
        id: bump
        uses: scouratier/km_build_upload@main
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_HUB_USERNAME }}
          password: ${{ secrets.DOCKER_HUB_ACCESS_TOKEN }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v6
        with:
          context: .
          file: ./Dockerfile
          push: true
          tags: |
            ${{ secrets.DOCKER_HUB_USERNAME }}/${{ steps.bump.outputs.image_name }}:latest
            ${{ secrets.DOCKER_HUB_USERNAME }}/${{ steps.bump.outputs.image_name }}:${{ steps.bump.outputs.new_version }}
```

### 3. Set up secrets

In your repository settings, add:
- `DOCKER_HUB_USERNAME` — Your Docker Hub username
- `DOCKER_HUB_ACCESS_TOKEN` — Your Docker Hub access token

## Action Outputs

| Output | Description | Example |
|---|---|---|
| `new_version` | The bumped semantic version | `1.2.4` |
| `image_name` | Docker image name (derived from repo name) | `km_brain_fe` |

## How It Works

1. Reads the `version` file (creates `0.0.0` if missing)
2. Bumps the patch version (e.g., `1.2.3` → `1.2.4`)
3. Writes the new version back to the `version` file
4. Commits with message `chore: bump version to X.Y.Z`
5. Creates git tag `X.Y.Z`
6. Pushes commit and tag to the remote
7. Outputs `new_version` and `image_name` for downstream steps

## Requirements

### Permissions

The workflow must have write access to push commits and tags. Add this to your workflow:

```yaml
permissions:
  contents: write
```

Alternatively, in your repository settings, enable **"Allow GitHub Actions to create and approve pull requests"** under Actions → General → Workflow permissions.

### Runner

This action requires a runner with Go pre-installed (included in `ubuntu-latest`).

## Version Format

- Versions follow semantic versioning: `major.minor.patch` (e.g., `1.2.3`)
- No `v` prefix — tags and the version file both use bare numbers
- Currently only patch bumping is supported

## Known Limitations & Tech Debt

- **Concurrent merges**: The action attempts to handle race conditions by pulling with `--rebase` before pushing. If multiple actions trigger simultaneously and modify the same files in conflicting ways, the rebase may fail. For highly active repositories, using [GitHub concurrency groups](https://docs.github.com/en/actions/using-jobs/using-concurrency) is recommended.
- **Patch only**: Currently only bumps the patch version. A future enhancement could scan commit messages for `#Minor` or `#Major` to control the bump level.
- **Dockerfile location**: Assumes the Dockerfile is always in the repository root.
- **Git identity**: Hardcoded to `github-actions[bot]`. Could be made configurable.

## Project Structure

```
km_build_upload/
├── action.yml              # Composite action definition
├── main.go                 # Go logic (version bump, file write, git tag)
├── main_test.go            # Unit tests
├── go.mod                  # Go module (stdlib only)
├── implementations/
│   ├── high_level.impl     # Original spec + review notes
│   └── docker-image.yml    # Reference workflow being replaced
├── examples/
│   └── workflow.yml        # Ready-to-use caller workflow template
├── PROJECT.md              # Project documentation
└── README.md               # This file
```

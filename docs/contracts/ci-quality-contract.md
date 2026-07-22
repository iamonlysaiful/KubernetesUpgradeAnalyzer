# CI and quality gates contract

Status: Phase 1 contract artifact
Last updated: 2026-07-23

This contract covers P1-03: the first continuous integration workflow and local
quality gate entrypoint.

## 1. Scope

P1-03 adds:

- a GitHub Actions workflow for pull requests and pushes to `main`;
- local scriptable quality gates for formatting, tests, vet, build, and JSON
  syntax checks;
- no live Kubernetes access;
- no Azure CLI invocation;
- no catalog download, web search, telemetry, release publishing, or deployment.

## 2. Workflow baseline

The MVP workflow uses GitHub-hosted Ubuntu runners and GitHub-owned setup
actions:

- `actions/checkout@v6`;
- `actions/setup-go@v6`.

Both are GitHub-maintained actions under the MIT license. The workflow grants
only `contents: read` permissions and does not use repository secrets.

`actions/setup-go` reads `go.mod` through `go-version-file: go.mod`. The current
module declares `go 1.25.0`; local development may use Go `1.26.x`.

## 3. Required gates

The first CI workflow runs:

- Go version display;
- `gofmt` check for all tracked Go files;
- `go test ./...`;
- `go vet ./...`;
- `go build -o /tmp/kua ./cmd/kua`;
- JSON syntax checks for all tracked schema and fixture JSON files.

Schema semantic validation, linting, race tests, SBOM generation, dependency
scanning, and release signing are deferred to later focused work packages after
their tooling and dependency choices are approved.

## 4. Local entrypoint

The local quality gate script is `scripts/ci-local.sh`. It must avoid writing
build outputs into the repository and must use `/tmp` or an explicitly supplied
temporary directory for generated binaries.

## 5. Failure policy

A failed quality gate blocks merge. A missing optional tool should fail with a
clear message rather than silently skipping a check. CI failures should be
reported with the failed command and whether the issue is code, environment, or
tooling.

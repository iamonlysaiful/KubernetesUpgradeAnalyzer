# Foundation implementation contract

Status: Phase 1 contract artifact
Last updated: 2026-07-22

This contract covers the first Go foundation slice before Kubernetes collectors,
provider adapters, catalog loading, or external process execution are added.

## 1. Scope

P1-01 establishes:

- module path `github.com/iamonlysaiful/KubernetesUpgradeAnalyzer`;
- binary name `kua`;
- standard-library CLI entrypoint and command dispatch;
- build-time version metadata placeholders;
- deterministic command names and exit-code constants;
- clear placeholders for unimplemented commands.

P1-01 does not add Kubernetes, Azure, kubent, Cobra, Viper, schema-validation,
lint, CI, or release dependencies. Those are handled by later focused work
packages with their own dependency assessments.

## 2. Toolchain baseline

The repository targets the current supported Go toolchain line. As of
2026-07-22, Go `1.26.5` is the latest stable patch release on go.dev. The module
declares `go 1.25.0` so it remains compatible with the currently supported prior
Go line while allowing Go `1.26.x` users to build normally.

Local verification requires `go` on PATH. If Go is absent, code may be reviewed
and committed, but build/test verification must be reported as not run.

## 3. Initial command behavior

| Command | MVP foundation behavior |
| --- | --- |
| `kua version` | Prints CLI name, version, commit, build date, Go runtime, schema version, and catalog version placeholders. |
| `kua analyze` | Returns a clear unimplemented message and exit code `5`. |
| `kua inventory` | Returns a clear unimplemented message and exit code `5`. |
| `kua health` | Returns a clear unimplemented message and exit code `5`. |
| `kua compatibility` | Returns a clear unimplemented message and exit code `5`. |
| `kua report` | Returns a clear unimplemented message and exit code `5`. |

Unknown commands and invalid usage return exit code `4`.

## 4. Exit-code constants

The foundation uses the accepted CLI contract:

- `0`: completed and `READY` or `READY_WITH_WARNINGS`;
- `2`: completed and `NOT_READY`;
- `3`: completed but `INCONCLUSIVE`;
- `4`: usage or configuration error;
- `5`: collection, dependency, catalog, internal execution, or unimplemented
  command error.

## 5. Dependency rule

The foundation intentionally starts without external modules. Adding Cobra,
Viper, `client-go`, schema tooling, lint tooling, or release tooling requires a
separate dependency assessment and user approval before the dependency is added
to `go.mod` or committed.

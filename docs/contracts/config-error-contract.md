# Config, logging, and error contract

Status: Phase 1 contract artifact
Last updated: 2026-07-22

This contract covers P1-02: the standard-library configuration, logging,
command-error, and exit-code foundation.

## 1. Scope

P1-02 adds:

- shared parsing for common global flags;
- a normalized runtime configuration object;
- log-level validation;
- command errors that carry user-facing messages and exit codes;
- deterministic mapping from command outcomes to the accepted CLI exit codes;
- unit tests for success, usage errors, unimplemented command errors, and
  logging/config validation.

P1-02 does not add Cobra, Viper, Kubernetes clients, Azure integration, file
config loading, catalog loading, report rendering, or live-system access.

## 2. Common flags accepted in P1-02

The foundation recognizes these global flags for every command:

| Flag | Values | P1-02 behavior |
| --- | --- | --- |
| `--log-level` | `debug`, `info`, `warn`, `error` | Validates and stores the requested level. |
| `--format` | `console`, `json`, `markdown`, `html` | Validates and stores the requested output format. |
| `--provider-source` | `auto`, `azure`, `file`, `offline`, `none` | Validates and stores the provider source without invoking providers. |
| `--context` | non-empty string | Stores the requested kubeconfig context without accessing Kubernetes. |
| `--kubeconfig` | non-empty string | Stores the requested kubeconfig path without reading it. |
| `--config` | non-empty string | Stores the future config-file path without reading it. |
| `--output` | non-empty string | Stores the future output path without writing it. |

Defaults are:

- `--log-level=info`
- `--format=console`
- `--provider-source=auto`

Flag parsing is intentionally minimal and supports `--flag value` and
`--flag=value`. Unknown flags and missing values return exit code `4`.

## 3. Logging boundary

P1-02 may use Go's standard `log/slog`, but it must not write logs for normal
`version` output. Future command diagnostics go to stderr. Log messages must not
include kubeconfig contents, credentials, provider tokens, raw cluster data, or
Secret data.

## 4. Error model

Commands return an application error type containing:

- machine category: usage, unimplemented, execution;
- user-facing message;
- exit code;
- optional cause for internal testing.

User-facing errors are concise and actionable. Normal command output goes to
stdout; errors and usage diagnostics go to stderr.

## 5. Exit-code mapping

P1-02 preserves the accepted exit-code contract:

- success: `0`;
- usage/configuration errors: `4`;
- unimplemented foundation commands: `5`;
- future `NOT_READY`: `2`;
- future `INCONCLUSIVE`: `3`.

No command may call `os.Exit` below `cmd/kua/main.go`; internal packages return
exit codes and errors for testability.

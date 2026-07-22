#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

go version

go_files="/tmp/kua-go-files.$$"
json_files="/tmp/kua-json-files.$$"
trap 'rm -f "$go_files" "$json_files"' EXIT

git ls-files -z '*.go' > "$go_files"
if [[ ! -s "$go_files" ]]; then
  echo "no tracked Go files found" >&2
  exit 1
fi

unformatted="$(xargs -0 gofmt -l < "$go_files")"
if [[ -n "$unformatted" ]]; then
  echo "gofmt required for:" >&2
  echo "$unformatted" >&2
  exit 1
fi

go test ./...
go vet ./...
go build -o /tmp/kua ./cmd/kua

git ls-files -z 'schemas/**/*.json' 'schemas/*.json' > "$json_files"
if [[ ! -s "$json_files" ]]; then
  echo "no tracked schema JSON files found" >&2
  exit 1
fi

xargs -0 jq empty < "$json_files"

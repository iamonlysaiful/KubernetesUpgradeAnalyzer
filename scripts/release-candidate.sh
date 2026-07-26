#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

version="${1:-}"
if [[ -z "$version" ]]; then
  version="0.1.0-rc.$(date -u +%Y%m%d%H%M%S)"
fi

commit="$(git rev-parse --short=12 HEAD)"
date_utc="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
go_version="$(go version | awk '{print $3}')"

out_dir="artifacts/release-candidate/${version}"
mkdir -p "$out_dir/bin"

platforms=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
)

for entry in "${platforms[@]}"; do
  os="${entry%% *}"
  arch="${entry##* }"
  out_name="kua_${version}_${os}_${arch}"
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w -X main.version=${version} -X main.commit=${commit} -X main.buildDate=${date_utc}" \
    -o "$out_dir/bin/$out_name" ./cmd/kua

done

(
  cd "$out_dir/bin"
  shasum -a 256 kua_* > ../SHA256SUMS
)

go list -m -json all > "$out_dir/SBOM-go-modules.json"

cat > "$out_dir/provenance.json" <<EOF
{
  "version": "${version}",
  "commit": "${commit}",
  "buildDate": "${date_utc}",
  "goVersion": "${go_version}",
  "schemaVersion": "kua.assessment.v1",
  "catalogVersion": "unavailable",
  "source": "$(git config --get remote.origin.url || echo unknown)"
}
EOF

cat > "$out_dir/RELEASE_NOTES.md" <<EOF
# KUA ${version} Release Candidate

## Build Metadata
- Commit: ${commit}
- Built at: ${date_utc}
- Go: ${go_version}

## Artifacts
- Binaries: ./bin
- Checksums: SHA256SUMS
- SBOM: SBOM-go-modules.json
- Provenance: provenance.json

## Validation
Run:
- scripts/ci-local.sh
- go test -race ./internal/recommendation ./internal/report

## Rollback
Releases are immutable. If a defect is found, publish a new patch candidate and mark this candidate deprecated.
EOF

echo "release candidate artifacts generated at: $out_dir"

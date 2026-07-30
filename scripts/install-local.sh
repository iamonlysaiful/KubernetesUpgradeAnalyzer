#!/usr/bin/env bash
# Build the current branch and install kua to /usr/local/bin for local verification.
# Run from the repository root. Does NOT publish anything.
set -euo pipefail

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="kua"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "Building kua from $(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')..."

VERSION="$(git describe --tags --exact-match 2>/dev/null || echo "0.0.0-local")"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"
BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

go build \
  -ldflags="-s -w \
    -X main.version=${VERSION} \
    -X main.commit=${COMMIT} \
    -X main.buildDate=${BUILD_DATE}" \
  -o "${ROOT}/bin/${BINARY_NAME}" \
  ./cmd/kua

echo "Build complete: ${ROOT}/bin/${BINARY_NAME}"
echo "  version: ${VERSION}  commit: ${COMMIT}"

echo
read -rp "Install to ${INSTALL_DIR}/${BINARY_NAME}? [y/N]: " answer
case "${answer,,}" in
  y|yes)
    if [[ -w "${INSTALL_DIR}" ]]; then
      cp "${ROOT}/bin/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
      sudo cp "${ROOT}/bin/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    echo "Installed: ${INSTALL_DIR}/${BINARY_NAME}"
    "${INSTALL_DIR}/${BINARY_NAME}" version
    ;;
  *)
    echo "Not installed. Binary is at ${ROOT}/bin/${BINARY_NAME}"
    ;;
esac

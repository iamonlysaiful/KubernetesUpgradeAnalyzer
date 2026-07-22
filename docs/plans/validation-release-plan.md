# Validation and release plan

Status: Approved plan; live validation and release require separate approval
Last updated: 2026-07-22

## 1. Validation ladder

1. Pure unit and schema tests.
2. Fake Kubernetes clients/discovery and controlled external-process fixtures.
3. Golden canonical/report fixtures.
4. Ephemeral local-cluster tests after dependency/execution approval.
5. Separately approved read-only AKS staging validation.
6. Release-candidate verification on Linux/macOS amd64/arm64.

No routine CI job receives production kubeconfig or Azure credentials.

## 2. AKS staging procedure

Before access, record approval, kube context, intended read-only commands, output location, redaction plan, and rollback/cleanup plan. Confirm context and identity, then run inventory, privacy review, health, kubent compatibility, provider evidence, recommendation, and reports in that order. Stop if collection exceeds the approved field/RBAC matrix.

Compare KUA with independent evidence:

- Kubernetes server/node versions and workload health;
- raw `az aks get-upgrades` JSON;
- kubent JSON with Helm collection disabled;
- manual component image/version inspection;
- documented API/component/provider support sources.

Expected initial scenario is current `1.30.0`, destination `1.33.12`, sequential intermediate minors, `READY`, and `LOW` only when all required evidence is sufficient.

## 3. Fixture handling

Never commit raw staging output. Create a recoverable local copy, sanitize through a reviewed field matrix, manually inspect for identifiers/secrets, and obtain explicit approval before committing any derived fixture. Preserve provenance without retaining sensitive values.

## 4. Release contents

- Versioned binaries for Linux/macOS amd64/arm64.
- Apache-2.0 license and notices.
- Embedded catalog version and source manifest.
- JSON Schemas and example redacted reports.
- SHA-256 checksums, SBOM, build provenance, changelog, known limitations, installation and verification instructions.

Signing keys, publishing credentials, and release automation require separate secure design/approval.

## 5. Release gates

- Clean working tree and healthy `git fsck`; no AppleDouble metadata in `.git`.
- Full formatting, unit, integration, golden, schema, lint, vet, race, security, secret, dependency, and license checks.
- Cross-platform build and smoke tests.
- Deterministic assessment and redaction equivalence checks.
- Catalog provenance/freshness review and kubent target-coverage decision.
- Approved staging validation record and known-limitations review.
- Explicit owner approval before tag, GitHub release, or artifact publication.

## 6. Rollback and recovery

Releases are immutable. A faulty release is deprecated, documented, and replaced with a new patch version; tags/assets/history are not silently rewritten. Preserve recoverable build inputs and previous catalog/tool versions. Any cleanup of validation or build artifacts follows the repository recoverable-cleanup procedure and retains backups until separately approved for deletion.

## 7. Versioning

Use semantic versioning. Pre-MVP builds remain `0.x`; schema/catalog versions are reported independently. Release notes list behavior, schema/catalog changes, supported Kubernetes/provider/tool ranges, security considerations, and upgrade instructions.

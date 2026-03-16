# Schema Contract Fixtures

This directory contains the shipped Epic 2 fixtures for schema validation, project-level extension-policy checks, and YAML-profile rejection cases in `v0.1.0-alpha.1`.

## Shipped Fixtures

The current fixture set contains 15 YAML fixtures:

### Valid standalone-schema fixtures

- `valid-runecontext-no-extensions.yaml`: root config with embedded source and no extensions.
- `valid-runecontext-with-extensions-optin.yaml`: root config with `allow_extensions: true` plus a valid namespaced `extensions` object.
- `valid-git-source-signed-tag.yaml`: root config using signed-tag git source details.
- `valid-bundle-closed-schema.yaml`: bundle with closed-schema fields only.
- `valid-bundle-with-extensions.yaml`: bundle using valid namespaced extension keys; requires project-level opt-in to be considered fully valid.
- `valid-change-status.yaml`: change status with standard lifecycle fields.
- `valid-custom-type.yaml`: change status using an `x-` custom type.
- `valid-superseded-change.yaml`: superseded change that includes `superseded_by`.
- `valid-context-pack.yaml`: generated context pack with git provenance and valid 64-character hashes.

### Reject standalone-schema fixtures

- `reject-unknown-field-runecontext.yaml`: unknown top-level field on `runecontext.yaml`.
- `reject-unknown-schema-version.yaml`: `schema_version: 2` must fail closed.
- `reject-bad-extension-key.yaml`: invalid extension key (`BadKey`) should fail namespacing rules.
- `reject-context-pack-unknown-field.yaml`: generated context pack with canonical four-aspect shape but forbidden unknown field `metadata`.

### Reject profile / project-level fixtures

- `reject-yaml-anchors-aliases.yaml`: YAML anchors/aliases violate the restricted machine-readable profile.
- `reject-extensions-without-optin.yaml`: change-status file with `extensions` but no accompanying root opt-in. This is a project-level validation case, not a standalone per-file schema failure.

## Validation Expectations

### Standalone JSON Schema validation

- All `valid-*.yaml` fixtures except `valid-bundle-with-extensions.yaml` validate directly against a single schema with no extra project context.
- `valid-bundle-with-extensions.yaml` validates structurally against `schemas/bundle.schema.json`; a full implementation must also confirm the root config enables extensions.
- All `reject-*.yaml` fixtures except `reject-extensions-without-optin.yaml` and `reject-yaml-anchors-aliases.yaml` should fail direct schema validation for the reason described in their filename and notes.

### Project-level validation

- `reject-extensions-without-optin.yaml` is expected to fail only when validation loads the project root and enforces the cross-file rule that bundle/status extensions require `runecontext.yaml` to set `allow_extensions: true`.
- `valid-bundle-with-extensions.yaml` likewise requires a root config with `allow_extensions: true` to be accepted in a real project.

### YAML profile validation

- `reject-yaml-anchors-aliases.yaml` is a parser/profile rejection case. Standard JSON Schema validation alone is not sufficient; implementations must reject it before or during YAML decoding according to `schemas/MACHINE-READABLE-PROFILE.md`.

## Coverage Notes

- Extension keys follow the enforced ownership pattern `owner.name.more`, allowing lowercase alphanumerics plus `_` and `-` within each non-empty segment while reserving `.` strictly as the namespace separator.
- Context-pack fixtures reflect the current provenance rules: only `git` sources may include `source_commit`; `embedded` and `path` sources must use their matching verification posture.
- Context-pack fixtures use the canonical artifact shape: `selected` always contains all four aspect keys, and `excluded` uses the same four-key layout whenever present.
- Context-pack hashes shown here are shape-valid placeholders for schema tests. Hash correctness against canonical JCS input should be covered by dedicated hashing tests in implementation code.

## Usage

1. Validate fixture syntax against the restricted YAML profile first.
2. Run standalone JSON Schema validation for the direct-schema cases.
3. Run project-level validation for extension opt-in behavior by loading both the root config and the bundle/status fixture together.
4. Reuse the same fixture expectations across Go and TypeScript implementations to preserve parity.

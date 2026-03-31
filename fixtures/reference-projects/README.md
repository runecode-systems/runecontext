# Reference Projects Fixtures

These fixtures cover the alpha.8 reference-project families used by install and
end-to-end validation coverage.

- `embedded/`: embedded-mode reference project.
- `linked-by-commit/`: linked git source fixture template for pinned commit.
- `linked-by-signed-tag/`: linked git source fixture template for signed-tag
  verification.
- `verified/`: verified-tier embedded reference project with baseline.
- `monorepo/`: nested RuneContext roots with root and service-level configs.

Linked fixture directories use `runecontext.yaml.tmpl` placeholders that tests
materialize with dynamic git repository values.

## Family Notes

- `embedded/` includes a complete local RuneContext tree and validates with
  explicit-root and CLI paths.
- `linked-by-commit/` and `linked-by-signed-tag/` intentionally hold only
  `runecontext.yaml.tmpl` because tests materialize those configs against
  temporary local git repositories (no network) built from
  `fixtures/source-resolution/templates/minimal-runecontext`.
- `verified/` mirrors embedded structure with `assurance_tier: verified` and a
  deterministic baseline fixture under `runecontext/assurance/`.
- `monorepo/` models nested RuneContext roots: top-level root config plus a
  service-local nested root under `packages/service/`.

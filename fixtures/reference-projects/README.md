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

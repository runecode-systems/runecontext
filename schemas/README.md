# Schemas

Hand-authored JSON Schema files for machine-readable RuneContext artifacts live here.

## v0.1.0-alpha.1 Schema Inventory

- `runecontext.schema.json`
  - root `runecontext.yaml` contract
- `bundle.schema.json`
  - `bundles/*.yaml` contract
- `change-status.schema.json`
  - `changes/*/status.yaml` contract
- `context-pack.schema.json`
  - generated context-pack contract
- `spec.schema.json`
  - YAML frontmatter contract for `specs/*.md`
- `decision.schema.json`
  - YAML frontmatter contract for `decisions/*.md`

## Notes

- All v1 schemas target JSON Schema Draft 2020-12.
- Machine-readable contracts stay closed by default and fail closed on unknown `schema_version` values.
- `specs/*.md` and `decisions/*.md` use YAML frontmatter as their strict traceability metadata layer; the markdown body remains hand-authored.
- The executable validation foundation in `internal/contracts/` compiles and exercises these schemas against repository fixtures.

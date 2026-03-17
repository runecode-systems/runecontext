# Traceability Fixtures

These fixtures exercise the strict alpha.1 traceability foundation for `changes/`, `specs/`, and `decisions/`.

- `valid-project/` is a self-contained embedded-mode project fixture with consistent cross-artifact references.
- `valid-project-custom-root/` proves whole-project validation follows `runecontext.yaml` source-root settings instead of assuming a fixed `runecontext/` directory.
- `reject-*` directories each model one fail-closed traceability violation.
- `specs/*.md` and `decisions/*.md` use YAML frontmatter with IDs that must match the path-relative stem.
- Project fixtures may also exercise root-config rules that affect hand-authored artifacts, such as extension opt-in behavior.

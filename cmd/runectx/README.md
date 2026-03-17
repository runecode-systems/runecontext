# runectx

This directory now contains the initial Go CLI entrypoint.

## Current Scope

- `runectx validate [path]`
  - validates the root `runecontext.yaml`
  - validates change status files, markdown contracts, spec/decision frontmatter,
    and project-level traceability
  - fails closed on schema, YAML-profile, and cross-artifact reference errors
  - emits stable line-oriented `key=value` output for success, validation
    failure, and usage errors without introducing full `--json` yet

## Output Contract

`runectx validate` uses a narrow machine-oriented output contract in alpha.1:

- success on stdout
  - `result=ok`
  - `command=validate`
  - `root=<absolute-path>`
- validation failure on stderr
  - `result=invalid`
  - `command=validate`
  - `root=<absolute-path>`
  - `error_path=<path>` when available
  - `error_message=<message>`
- usage/command errors on stderr
  - `result=usage_error`
  - `command=<name>` when available
  - `error_message=<message>`
  - `usage=runectx validate [path]`

Backslashes and newlines in values are escaped so each field stays on one line.
Consumers should split each output record on the first `=` character.

## Current Constraint

- In the current repository-first alpha.1 implementation, schemas are discovered
  relative to the source checkout. Release-oriented schema embedding or explicit
  schema-root overrides remain future work.
- Whole-project validation follows the content root declared in `runecontext.yaml`
  (`source.path` for embedded/path sources and `source.subdir` for git sources)
  rather than assuming a fixed `runecontext/` directory.

The CLI remains intentionally narrow in alpha.1. Broader command coverage is
still planned for later milestones.

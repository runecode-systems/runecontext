# runectx

This directory now contains the initial Go CLI entrypoint.

## Current Scope

- `runectx status [path]`
  - reports active, closed, and superseded changes plus bundle and root metadata
- `runectx change new --title TITLE --type TYPE ...`
  - creates a minimum change by default and auto-shapes large or project work
  - writes `status.yaml`, `proposal.md`, and `standards.md`
- `runectx change shape CHANGE_ID ...`
  - materializes `design.md` and `verification.md` by default
  - creates `tasks.md` and `references.md` only when non-empty content is provided
  - fails closed for terminal changes rather than mutating historical artifacts
- `runectx change close CHANGE_ID ...`
  - closes or supersedes a change without moving it off its stable path
  - fails closed if a missing reciprocal `supersedes` link would require mutating
    an already-terminal successor change
- `runectx validate [path]`
  - validates the root `runecontext.yaml`
  - validates change status files, markdown contracts, spec/decision frontmatter,
    and project-level traceability
  - fails closed on schema, YAML-profile, and cross-artifact reference errors
  - emits stable line-oriented `key=value` output for success, validation
    failure, and usage errors without introducing full `--json` yet

## Output Contract

The alpha.3 thin commands all use the same narrow line-oriented machine contract:

- success on stdout
  - `result=ok`
  - `command=<command>`
  - `root=<absolute-path>`
  - command-specific metadata such as change IDs, lifecycle counts, standards
    refresh details, and changed file paths/actions
- validation failure on stderr
  - `result=invalid`
  - `command=<command>`
  - `root=<absolute-path>`
  - `error_path=<path>` when available
  - `error_message=<message>`
- usage/command errors on stderr
  - `result=usage_error`
  - `command=<name>` when available
  - `error_message=<message>`
  - `usage=<command-usage>`

Values escape `\\`, `\n`, `\r`, `\t`, `\0`, and `\=` so each field stays on
one line. Consumers should split each output record on the first `=`
character and then reverse those escape sequences.

## Current Constraint

- In the current repository-first alpha.3 implementation, schemas are discovered
  relative to the source checkout. Release-oriented schema embedding or explicit
  schema-root overrides remain future work.
- Whole-project validation follows the content root declared in `runecontext.yaml`
  (`source.path` for embedded/path sources and `source.subdir` for git sources)
  rather than assuming a fixed `runecontext/` directory.
- Explicit path arguments are respected as explicit roots even when the caller
  passes `.`; the CLI does not silently climb to an ancestor RuneContext root in
  that case.

The CLI remains intentionally narrow in alpha.3. Broader command coverage is
still planned for later milestones.

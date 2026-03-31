---
name: runecontext
description: RuneContext discoverability shim index
---
<!-- runecontext-managed-artifact: host-native-v1 -->
<!-- runecontext-tool: claude-code -->
<!-- runecontext-kind: discoverability_shim -->
<!-- runecontext-id: runecontext:index -->
# RuneContext Command Shim Index

This file is a discoverability shim. Canonical flow assets live in `.claude/skills/`.

- Adapter role: discoverability shim

## Commands

- `runecontext:change-new`
  - Canonical flow source: `adapters/claude-code/flows/change-new.md`
  - Skill file: `.claude/skills/runecontext-change-new.md`

- command_path: `change new`

- `runecontext:change-shape`
  - Canonical flow source: `adapters/claude-code/flows/change-shape.md`
  - Skill file: `.claude/skills/runecontext-change-shape.md`

- command_path: `change shape`

- `runecontext:standard-discover`
  - Canonical flow source: `adapters/claude-code/flows/standard-discover.md`
  - Skill file: `.claude/skills/runecontext-standard-discover.md`

- command_path: `standard discover`

- `runecontext:promote`
  - Canonical flow source: `adapters/claude-code/flows/promote.md`
  - Skill file: `.claude/skills/runecontext-promote.md`

- command_path: `promote`

!`runectx adapter render-host-native --role discoverability_shim claude-code index`

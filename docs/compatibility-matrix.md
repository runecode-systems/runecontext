# RuneContext ↔ RuneCode Compatibility Matrix

RuneCode uses the `runecontext_version` field in `runecontext.yaml` as the
compatibility gate for project upgrades and runtime wiring.

## Version compatibility

| RuneCode release | Acceptable `runecontext_version` range | Adapter-pack compatibility | Notes |
| --- | --- | --- | --- |
| `v0.1.0-alpha.*` | `0.1.0-alpha.5` – `0.1.0-alpha.8` | `adapter-generic`, `adapter-codex`, `adapter-claude-code`, `adapter-opencode` (same release tag) | Alpha flows remain repo-first; RuneCode should require matching release-line assets from the same signed release set. |
| `v0.1.0` | `0.1.0` (planned GA) | Same adapter packs plus any GA follow-up adapters | At GA, compatibility freezes to the GA contract and older alpha `runecontext_version` values should be rejected. |

If a project reports an out-of-range `runecontext_version`, validation and
integration flows should fail closed and direct users to upgrade.

## Upgrade preview and migration-path semantics

- `runectx upgrade` is a read-only assessment/dry-run surface.
- `runectx upgrade apply` remains the only mutating upgrade command.
- `runectx upgrade cli` is the explicit CLI self-update preview/check surface.
- `runectx upgrade cli apply` is the explicit mutating CLI self-update surface.
- Upgrade planning follows explicit registered migration hops. Preview reports an
  ordered hop chain (`hop_count`, `hop_N_from`, `hop_N_to`) plus readable
  per-hop actions.
- If no registered migration path exists but the selected target only requires a
  compatible version bump with no migration logic, planning may still produce a
  zero-hop version-bump-only upgrade.
- If the requested target requires migrations and no registered path exists,
  planning fails closed with an unsupported-project-version state instead of
  auto-bumping the version.

## Release distribution semantics

RuneContext release semantics are intentionally two-lane:

- **canonical release path**: verify and consume the repo bundle assets produced
  by `nix build --no-link .#release-artifacts` (`runecontext_<tag>.tar.gz` or
  `runecontext_<tag>.zip`), plus `schema-bundle.tar.gz` and adapter packs.
- **optional binary convenience path**: verify and install
  `runecontext_<tag>_<os>_<arch>.tar.gz` when users only need the `runectx`
  executable.

Both lanes are tied together by the same release manifest, checksum set,
signature/certificate set, and provenance attestation set.

### Optional binary platform matrix

| OS | Arch | Archive asset |
| --- | --- | --- |
| `linux` | `amd64` | `runecontext_<tag>_linux_amd64.tar.gz` |
| `linux` | `arm64` | `runecontext_<tag>_linux_arm64.tar.gz` |
| `darwin` | `amd64` | `runecontext_<tag>_darwin_amd64.tar.gz` |
| `darwin` | `arm64` | `runecontext_<tag>_darwin_arm64.tar.gz` |

Windows remains repo-bundle-only for now and is not part of the optional binary
asset set.

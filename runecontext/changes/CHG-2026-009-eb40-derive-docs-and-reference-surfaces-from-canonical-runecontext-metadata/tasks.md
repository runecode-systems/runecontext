# Tasks

- Define the generated reference artifact format for Docus consumption, favoring structured JSON over hand-maintained markdown copies.
- Generate command, capability, compatibility, and layout/reference data from the canonical metadata builder so website reference pages can render from stable IDs and tokens.
- Replace or redirect stale references that still claim `runecontext/operations/` is the canonical in-repo operations reference.
- Add a unified metadata sync path so docs/reference generation and related metadata-derived artifacts refresh together from `nix/release/metadata.nix`.
- Add parity checks so docs/reference generation fails when CLI registry or capability metadata changes without corresponding regenerated outputs.

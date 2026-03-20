# Machine-Readable YAML Profile

This document defines the restricted YAML profile used for all machine-readable RuneContext artifacts to ensure deterministic hashing and cross-implementation compatibility.

## Scope

This profile applies to:
- `runecontext.yaml` (root configuration)
- `bundles/*.yaml` (context selectors)
- `changes/*/status.yaml` (change lifecycle)
- All generated artifacts: `context-pack.yaml`, `assurance/baseline.yaml`, receipt files

## JSON Schema Dialect

- All JSON schemas in `schemas/*.schema.json` use JSON Schema Draft 2020-12.
- Draft 2020-12 is required so closed contracts can safely combine conditional branches with `unevaluatedProperties: false`.
- Implementations should standardize on validators that fully support Draft 2020-12 in every language runtime they ship.

## Restricted YAML Syntax

### Required Constraints

- **No anchors or aliases**: YAML anchors (`&anchor`) and aliases (`*anchor`, `<<`) are forbidden. Each value must be written out in full.
- **No duplicate keys**: Objects must not contain duplicate keys.
- **No implicit type coercion beyond schema**: Values must match the schema's intended type. For example, `yes`/`no` must not be coerced to booleans; use `true`/`false` explicitly.
- **No custom tags**: YAML tags like `!!str`, `!!int`, `!!timestamp` are forbidden.
- **UTF-8 only**: All files must be valid UTF-8. No other encodings.
- **No null keys or values outside schema**: Null values (`null`) are allowed only where the schema explicitly permits them.
- **Normalized formatting**: Generated artifacts must use consistent indentation (2-space), no trailing whitespace, Unix line endings (LF), and end with a single newline.

### Collections and Nesting

- Arrays (`- item`) and objects (`key: value`) follow standard YAML nesting rules.
- Empty arrays and objects are permitted where the schema allows them.
- No flow-style collections: Use block-style (`- item`, `key: value`) exclusively.

### Strings

- Plain scalars, single-quoted, and double-quoted strings are all acceptable.
- No multiline strings with `|` or `>` syntax.
- String values must not contain unescaped control characters.

## Canonical Data Model For Hashing

### Transformation From YAML To JSON

To compute deterministic hashes (using RFC 8785 JCS), machine-readable artifacts must be transformed to a normalized JSON model as follows:

1. **Parse the YAML** into memory as a nested structure of objects, arrays, strings, numbers, booleans, and nulls.
2. **Normalize strings**: All strings remain strings; no implicit type conversions.
3. **Sort object keys**: All object keys must be sorted lexicographically.
4. **Remove whitespace**: Convert to compact JSON with no unnecessary spaces.
5. **Serialize to JSON**: Use standard JSON syntax (`"key": value`, `[...]`, `{...}`).
6. **Apply RFC 8785 JCS canonicalization**: See "JCS Canonicalization" below.

### Example: Source vs. Canonical

**Source YAML:**
```yaml
schema_version: 1
id: my-bundle
includes:
  standards:
    - standards/global/**
  project:
    - project/mission.md
```

**Canonical JSON (before JCS):**
```json
{
  "id": "my-bundle",
  "includes": {
    "project": ["project/mission.md"],
    "standards": ["standards/global/**"]
  },
  "schema_version": 1
}
```

(Keys are sorted lexicographically: `id`, `includes`, `schema_version`.)

## JCS Canonicalization (RFC 8785)

For hash computation, use RFC 8785 JSON Canonicalization Scheme (JCS):

1. **Object keys**: Sort lexicographically. Duplicate keys are forbidden by the profile anyway.
2. **Numbers**: Serialize without leading zeros, no unnecessary decimals. Example: `1` not `1.0`; `0.5` not `.5`.
3. **Whitespace**: No whitespace around separators (`:`, `,`).
4. **Unicode escape sequences**: Use minimally; prefer UTF-8 encoding of characters.
5. **Floating-point precision**: Avoid; use integers or strings for precision-sensitive values.

### Hash Algorithm And Output

- **Algorithm**: SHA256
- **Input**: UTF-8 bytes of JCS canonical JSON
- **Output**: Lowercase hexadecimal (64 characters)

Example:
```
Input (JCS): {"id":"my-bundle","schema_version":1}
SHA256 hash: a1b2c3d4... (64 hex chars)
```

### Context Pack Hash Input

- Context packs currently use the explicit canonicalization token `runecontext-canonical-json-v1` rather than claiming full RFC 8785 JCS interoperability.
- `runecontext-canonical-json-v1` is a restricted canonical JSON profile for the value shapes emitted by alpha.4 context packs: sorted object keys, compact arrays/objects, standard JSON string escaping without HTML escaping, and integral numeric values only. Other value shapes must fail closed rather than being serialized approximately.
- Strings participating in `runecontext-canonical-json-v1` must be valid UTF-8; canonicalization must fail closed on invalid UTF-8 instead of silently replacing bytes with U+FFFD or another normalization artifact.
- Context packs must exclude the `pack_hash` field itself before canonicalizing the remaining object for hashing.
- `generated_at` remains a required emitted field for context packs, but it is excluded from the canonical hash input so identical resolved content hashes the same across regenerations.
- The canonical hash input is the full context-pack object containing exactly these top-level fields when present: `schema_version`, `canonicalization`, `pack_hash_alg`, `id`, `requested_bundle_ids`, `resolved_from`, `selected`, and `excluded`.
- `selected` must always serialize with all four aspect keys: `project`, `standards`, `specs`, and `decisions`, using empty arrays when an aspect selects no files.
- `excluded` must always be present and serialize with the same four aspect keys, using empty arrays for aspects with no excluded files.
- `resolved_from`, `selected`, and `excluded` contribute their full nested content exactly as stored in the pack.
- `generated_at` must be supplied explicitly to the core builder as a whole-second UTC timestamp; builders must reject sub-second values instead of silently truncating them.
- For selected-file `sha256` values, UTF-8 text files are normalized to LF line endings before hashing; non-UTF-8 or binary files are hashed as raw bytes.
- When `resolved_from.source_mode` is `path`, `source_ref` must be a portable forward-slash relative path with no drive-qualified, UNC, `.` or `..` segments.
- Run `runecontext-canonical-json-v1` over that truncated object, then compute SHA256 over the UTF-8 bytes; the resulting 64-character hex string becomes the value stored in `pack_hash`.

## Unknown-Field Behavior

### Closed Schemas (Default)

By default, machine-readable artifacts use closed JSON schemas:
- Unknown top-level fields are rejected during validation.
- No arbitrary additions are allowed.

### Extensions Mechanism (Opt-In)

Hand-authored files (`runecontext.yaml`, `bundles/*.yaml`, `changes/*/status.yaml`) may include an optional `extensions` object when explicitly enabled:

1. **Opt-in requirement**: The project must set `allow_extensions: true` in `runecontext.yaml`.
2. **Validation scope**: Root-level `runecontext.yaml` can enforce this directly in its own schema. Bundle/status files require project-level validation because the opt-in flag lives in a different file.
3. **When extensions present**: Validation passes with a warning once the project-level opt-in check succeeds.
4. **Namespaced keys**: Extension keys must follow ownership-style namespacing:
   - Format: `[a-z0-9](?:[a-z0-9_-]*[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9_-]*[a-z0-9])?)+` (e.g., `dev.acme.foo`, `io.runecode.custom_metadata`).
   - Each segment may include underscores or dashes, but dots are reserved strictly as namespace separators so empty segments like `dev..meta` fail validation.
   - Prevents collisions, surfaces typos early, and keeps extension keys auditable.
5. **Non-authoritative**: Extension values are data, not semantics. They cannot affect:
   - Schema validation outcomes
   - Bundle resolution behavior
   - Change lifecycle states
   - Context pack generation
   - Assurance tier meaning
   - Policy or approvals
6. **Included in hashes**: Extension data is part of the YAML content, so changes to `extensions` affect the file's hash.

**Extensions are NOT permitted in generated artifacts** (context packs, baselines, receipts).

## Validation Contract

- Implementations must reject unknown `schema_version` values.
- Implementations must reject unknown values for enum fields (e.g., `status`, `type`, `source_verification`).
- Implementations must enforce all required fields per the schema.
- Implementations must reject files that violate the restricted YAML profile (anchors, aliases, duplicate keys, custom tags, non-UTF-8).
- Implementations must preserve the YAML structure when round-tripping if the input is valid.
- Context packs may omit `source_commit` when the source is not `git`; implementations must enforce the requirement only for `source_mode: git` to avoid forcing synthetic hashes on embedded or local resolutions.
- Implementations must enforce source-mode and verification consistency for context packs: `embedded` sources use `source_verification: embedded`, `path` sources use `source_verification: unverified_local_source`, and only `git` sources may record `source_commit`.

## Cross-Implementation Compatibility

To ensure local and remote resolution produce identical results:

1. **Use this profile for all authoritative YAML/JSON files**.
2. **Use the documented canonicalization profile for each artifact family**: RFC 8785 JCS remains the default profile, while alpha.4 context packs explicitly use `runecontext-canonical-json-v1` until a broader full-JCS implementation is shipped for that artifact.
3. **Test parity fixtures across implementations** (Go, TypeScript, etc.), including project-level validation that checks bundle/status extensions against the root opt-in flag.
4. **Document any deviations** from this profile in implementation release notes.

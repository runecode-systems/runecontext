## Summary
Allow local path source resolution for developer-local workflows.

## Problem
Developers need a non-auditable but convenient way to point at a local RuneContext tree during iteration.

## Proposed Change
Resolve local `type: path` sources only in explicit local mode and record warnings in structured metadata.

## Why Now
Alpha.2 needs the storage modes to behave consistently before later hashing work lands.

## Assumptions
The local tree can be snapshotted during resolution for future hashing and TOCTOU hardening.

## Out of Scope
Any audited or verified treatment of local path sources.

## Impact
Developers can test local RuneContext content without implying reproducible provenance.

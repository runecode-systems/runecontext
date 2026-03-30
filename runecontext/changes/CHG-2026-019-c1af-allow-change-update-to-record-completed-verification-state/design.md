# Design

## Overview
Extend change update so a non-terminal lifecycle mutation can also record a completed verification status when that state is required by validation. The command should let users move an open change to status verified without forcing terminal close, provided verification_status is explicitly set to passed, failed, or skipped in the same mutation or is already completed. Validation should continue rejecting contradictory states such as verified plus pending verification_status. Lifecycle progression remains non-terminal: change update must not write closed or superseded, and promotion_assessment remains unchanged.

## Command Rules
- `change update` should accept an explicit verification-status input for non-terminal lifecycle updates.
- The command should allow `status=verified` only when `verification_status` is already completed or is set to `passed`, `failed`, or `skipped` by the same mutation.
- Other non-terminal lifecycle updates may leave `verification_status` unchanged unless the user explicitly sets it.

## Mutation Rules
- `promotion_assessment` remains untouched.
- The command must continue rejecting terminal lifecycle writes.
- Backward lifecycle transitions remain invalid.

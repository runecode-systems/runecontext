# Tasks

- Add explicit verification-status mutation support to change update for non-terminal lifecycle transitions that require completed verification metadata.
- Allow status verified only when verification_status is already completed or is being set to passed, failed, or skipped in the same update operation.
- Preserve the non-terminal contract: update must not close or supersede changes and must not mutate promotion_assessment.

# Tasks

- Add an explicit recursive flag for change update and change close that enables umbrella-to-sub-change lifecycle propagation.
- Define the recursive target set narrowly: only feature sub-changes associated with the selected project umbrella, never arbitrary related_changes entries.
- Make recursive mutation transactional and fail closed if any targeted sub-change cannot legally take the requested lifecycle or verification transition.
- Expose recursive intent and affected change counts in command output so users can review the cascade they requested.

# Tasks

- Implement ./bin/runectx change update <CHANGE_ID> --status planned|implemented|verified --path PATH with stable JSON and human output contracts.
- Ensure the command validates lifecycle ordering, rejects invalid backward or terminal transitions, and leaves promotion_assessment untouched for non-terminal updates.
- Document and test the intended workflow split between change lifecycle advancement and close-time promotion assessment.

# Generic Adapter: Manual Flow

Manual, review-first workflow without host automation.

1. Inspect status: `runectx status --path <project-root>`
2. Discover standards candidates: `runectx standard discover --path <project-root>`
3. Create change: `runectx change new --title "..." --type feature --path <project-root>`
4. Shape change as needed: `runectx change shape <CHANGE_ID> --path <project-root>`
5. Validate authored content: `runectx validate --path <project-root>`

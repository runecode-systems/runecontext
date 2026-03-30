package cli

import "path/filepath"

func metadataSchemaPathFromRepoRoot(root string) string {
	return filepath.Join(root, "schemas", metadataSchemaName)
}

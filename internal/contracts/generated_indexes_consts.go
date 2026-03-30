package contracts

const (
	generatedManifestSchemaVersion    = 1
	generatedChangeIndexSchemaVersion = 1
	generatedBundleIndexSchemaVersion = 1
	generatedManifestRelativePath     = "manifest.yaml"
	generatedChangesIndexRelativePath = "indexes/changes-by-status.yaml"
	generatedBundlesIndexRelativePath = "indexes/bundles.yaml"
	generatedIndexesDirectoryRelative = "indexes"
)

// GeneratedManifestRelativePathForCLI exposes the canonical relative manifest
// output path used by generated artifact contracts.
func GeneratedManifestRelativePathForCLI() string {
	return generatedManifestRelativePath
}

// GeneratedIndexesDirectoryRelativeForCLI exposes the canonical relative
// indexes root path used by generated artifact contracts.
func GeneratedIndexesDirectoryRelativeForCLI() string {
	return generatedIndexesDirectoryRelative
}

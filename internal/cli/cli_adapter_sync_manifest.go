package cli

import (
	"fmt"
	"strings"
)

func buildAdapterManifest(tool string, managedFiles []string, hostNativeFiles []hostNativeArtifact) []byte {
	lines := []string{
		"schema_version: 1",
		"adapter: " + tool,
		"source: local_release",
		"manifest_kind: convenience_metadata",
		fmt.Sprintf("managed_file_count: %d", len(managedFiles)),
		"managed_files:",
	}
	for _, rel := range managedFiles {
		lines = append(lines, "  - managed/"+rel)
	}
	flowAssets, shims, hostNativePaths := hostNativeManifestSections(hostNativeFiles)
	lines = append(lines,
		fmt.Sprintf("host_native_file_count: %d", len(hostNativeFiles)),
		"host_native_files:",
	)
	for _, rel := range hostNativePaths {
		lines = append(lines, "  - "+rel)
	}
	lines = append(lines, "host_native_flow_assets:")
	for _, rel := range flowAssets {
		lines = append(lines, "  - "+rel)
	}
	lines = append(lines, "host_native_discoverability_shims:")
	for _, rel := range shims {
		lines = append(lines, "  - "+rel)
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func hostNativeManifestSections(hostNativeFiles []hostNativeArtifact) ([]string, []string, []string) {
	flowAssets := make([]string, 0)
	shims := make([]string, 0)
	hostNativePaths := make([]string, 0, len(hostNativeFiles))
	for _, artifact := range hostNativeFiles {
		hostNativePaths = append(hostNativePaths, artifact.relPath)
		if artifact.shim {
			shims = append(shims, artifact.relPath)
			continue
		}
		flowAssets = append(flowAssets, artifact.relPath)
	}
	return flowAssets, shims, hostNativePaths
}

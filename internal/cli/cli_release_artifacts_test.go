package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestReleaseMetadataDeclaresSchemaBundleAndAdapterPacks(t *testing.T) {
	metadata := readReleaseFileForTests(t, filepath.Join("nix", "release", "metadata.nix"))

	requireSubstrings(t, metadata,
		`name = "schema-bundle";`,
		`entries = [`,
		`"schemas"`,
		`name = "adapter-generic";`,
		`name = "adapter-codex";`,
		`name = "adapter-claude-code";`,
		`name = "adapter-opencode";`,
	)
}

func TestReleaseMetadataDeclaresOptionalBinaryTargets(t *testing.T) {
	metadata := readReleaseFileForTests(t, filepath.Join("nix", "release", "metadata.nix"))

	want := []string{
		"darwin/amd64/tar.gz",
		"darwin/arm64/tar.gz",
		"linux/amd64/tar.gz",
		"linux/arm64/tar.gz",
	}
	got := parseReleaseTargets(t, metadata)
	if !equalStrings(got, want) {
		t.Fatalf("unexpected release targets: got %v, want %v", got, want)
	}
}

func TestReleaseArtifactBuilderRecordsManifestAndChecksumCoverage(t *testing.T) {
	script := readReleaseFileForTests(t, filepath.Join("nix", "scripts", "build-release-artifacts.sh"))

	requireSubstrings(t, script,
		`process_pack_archives "schema_bundle"`,
		`process_pack_archives "adapter_pack"`,
		`"${coreutils}/cp" -R schemas "${share_dir}/schemas"`,
		`"${coreutils}/cp" -R adapters "${share_dir}/adapters"`,
		`"${host_metadata_binary}" metadata > "release/metadata-descriptor.json"`,
		`if ! @jq@/bin/jq -e '`,
		`invalid metadata descriptor payload: release/metadata-descriptor.json`,
		`record_archive "installer_script"`,
		`record_archive "repo_bundle"`,
		`record_archive "binary"`,
		`manifest_path="release/dist/@packageName@_@tag@_release-manifest.json"`,
		`metadata_descriptor: $metadata_descriptor[0],`,
		`release_files=( *.tar.gz *.zip *.json *.sh *.ps1 )`,
	)
}

func TestInstallScriptsInstallRuntimeAssets(t *testing.T) {
	sh := readReleaseFileForTests(t, filepath.Join("scripts", "install-runectx.sh"))
	requireSubstrings(t, sh,
		`runtime_source="${package_dir}/share/runecontext"`,
		`runtime_target="${install_prefix}/share/runecontext"`,
		`cp -R "${runtime_source}" "${runtime_target}"`,
	)

	ps1 := readReleaseFileForTests(t, filepath.Join("scripts", "install-runectx.ps1"))
	requireSubstrings(t, ps1,
		`$runtimeSource = Join-Path $packageDir "share/runecontext"`,
		`$runtimeTarget = Join-Path $installPrefix "share/runecontext"`,
		`Copy-Item -Path $runtimeSource -Destination $runtimeTarget -Recurse -Force`,
	)
}

func TestReleaseWorkflowUsesManifestDrivenAssetSetAndEnvironmentGate(t *testing.T) {
	workflow := readReleaseFileForTests(t, filepath.Join(".github", "workflows", "release.yml"))

	requireSubstrings(t, workflow,
		"name: release",
		"nix build --no-link --print-out-paths --no-write-lock-file .#release-artifacts",
		`mapfile -t archive_assets < <(jq -er '.archives[].file' "${manifest}")`,
		`required_assets+=("${archive_assets[@]}")`,
		`for suffix in "" ".sig" ".pem"; do`,
		"release/dist/*.tar.gz",
		"release/dist/*.zip",
		"release/dist/*.json",
		"release/dist/*.sh",
		"release/dist/*.ps1",
		"install-runectx.sh",
		"install-runectx.ps1",
		"release/dist/SHA256SUMS",
	)
}

func TestCompatibilityMatrixDocumentsCanonicalAndOptionalReleasePaths(t *testing.T) {
	matrix := readReleaseFileForTests(t, filepath.Join("docs", "compatibility-matrix.md"))

	requireSubstrings(t, matrix,
		"canonical release path",
		"optional binary convenience path",
		"`nix build --no-link .#release-artifacts`",
		"`runecontext_<tag>.tar.gz`",
		"`runecontext_<tag>_<os>_<arch>.tar.gz`",
		"| `linux` | `amd64` | `runecontext_<tag>_linux_amd64.tar.gz` |",
		"| `linux` | `arm64` | `runecontext_<tag>_linux_arm64.tar.gz` |",
		"| `darwin` | `amd64` | `runecontext_<tag>_darwin_amd64.tar.gz` |",
		"| `darwin` | `arm64` | `runecontext_<tag>_darwin_arm64.tar.gz` |",
	)
}

func TestReleaseManifestMetadataDescriptorParityFixture(t *testing.T) {
	withReleaseMetadataVersionForTests(t, func() {
		assertReleaseManifestDescriptorParityFixture(t)
	})
}

func assertReleaseManifestDescriptorParityFixture(t *testing.T) {
	t.Helper()
	root, err := repoRootForTests()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	manifestPath := filepath.Join(root, "fixtures", "release", "release-manifest-with-metadata.json")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read fixture release manifest: %v", err)
	}
	manifestDescriptor, err := releaseManifestDescriptorFromJSON(raw)
	if err != nil {
		t.Fatalf("parse fixture metadata descriptor: %v", err)
	}
	expected, err := descriptorMap(buildCapabilityDescriptor())
	if err != nil {
		t.Fatalf("build expected descriptor map: %v", err)
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("marshal expected descriptor: %v", err)
	}
	manifestJSON, err := json.Marshal(manifestDescriptor)
	if err != nil {
		t.Fatalf("marshal manifest descriptor: %v", err)
	}
	if string(manifestJSON) != string(expectedJSON) {
		t.Fatalf("expected release manifest fixture metadata_descriptor to match canonical descriptor\nexpected=%s\nactual=%s", string(expectedJSON), string(manifestJSON))
	}
}

func readReleaseFileForTests(t *testing.T, relativePath string) string {
	t.Helper()
	root, err := repoRootForTests()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, relativePath))
	if err != nil {
		t.Fatalf("read %s: %v", relativePath, err)
	}
	return string(content)
}

func requireSubstrings(t *testing.T, body string, expected ...string) {
	t.Helper()
	for _, fragment := range expected {
		if strings.Contains(body, fragment) {
			continue
		}
		t.Fatalf("expected content to include %q", fragment)
	}
}

func parseReleaseTargets(t *testing.T, metadata string) []string {
	t.Helper()
	pattern := regexp.MustCompile(`(?s)\{\s*goos = "([^"]+)";\s*goarch = "([^"]+)";\s*archive = "([^"]+)";\s*\}`)
	matches := pattern.FindAllStringSubmatch(metadata, -1)
	if len(matches) == 0 {
		t.Fatal("expected at least one target in release metadata")
	}
	targets := make([]string, 0, len(matches))
	for _, match := range matches {
		targets = append(targets, match[1]+"/"+match[2]+"/"+match[3])
	}
	sort.Strings(targets)
	return targets
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

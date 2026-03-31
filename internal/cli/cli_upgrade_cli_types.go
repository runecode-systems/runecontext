package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type cliUpgradeRequest struct {
	targetVersion string
	apply         bool
}

type cliUpgradeAvailability string

const (
	cliUpgradeAvailabilityUpToDate  cliUpgradeAvailability = "up_to_date"
	cliUpgradeAvailabilityAvailable cliUpgradeAvailability = "update_available"
	cliUpgradeAvailabilityUnknown   cliUpgradeAvailability = "unknown"
)

type cliUpgradePlan struct {
	Availability      cliUpgradeAvailability
	CurrentVersion    string
	SelectedRelease   string
	TargetRelease     string
	RequestedTarget   string
	NetworkAccess     bool
	Mutating          bool
	PlannedAction     string
	FailureGuidance   []string
	ReleaseSource     string
	Platform          string
	InstallAction     string
	InstalledBinary   string
	UpdatedBinaryPath string
	Changed           bool
}

type cliUpgradeResolver interface {
	ResolveLatestRelease(currentVersion string) (string, error)
}

type cliUpgradeInstaller interface {
	Apply(plan cliUpgradePlan) (cliUpgradePlan, error)
}

var resolveLatestCLIReleaseFn cliUpgradeResolver = staticLatestCLIReleaseResolver{}
var applyCLIUpgradePlanFn cliUpgradeInstaller = defaultCLIUpgradeInstaller{}
var cliUpgradeExecutablePathFn = os.Executable

type staticLatestCLIReleaseResolver struct{}

func (staticLatestCLIReleaseResolver) ResolveLatestRelease(currentVersion string) (string, error) {
	root, err := cliUpgradeRuntimeRoot()
	if err != nil {
		return "", fmt.Errorf("resolve latest release metadata root: %w", err)
	}
	version, err := readLatestCLIReleaseFromRuntimeRoot(root)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(version) == "" {
		return "", fmt.Errorf("latest release version is empty")
	}
	return strings.TrimPrefix(version, "v"), nil
}

func findRepoRootForReleaseMetadata(start string) (string, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		metadataPath := filepath.Join(abs, releaseMetadataRelativePath)
		if _, statErr := os.Stat(metadataPath); statErr == nil {
			return abs, nil
		}
		next := filepath.Dir(abs)
		if next == abs {
			return "", fmt.Errorf("could not locate %s from %s", releaseMetadataRelativePath, start)
		}
		abs = next
	}
}

func cliUpgradeRuntimeRoot() (string, error) {
	return locateCLIUpgradeRuntimeRoot(cliUpgradeRuntimeDeps{executable: cliUpgradeExecutablePathFn})
}

type cliUpgradeRuntimeDeps struct {
	executable func() (string, error)
}

func locateCLIUpgradeRuntimeRoot(deps cliUpgradeRuntimeDeps) (string, error) {
	deps = normalizeCLIUpgradeRuntimeDeps(deps)
	starts := cliUpgradeRuntimeStartPaths(deps)
	seen := map[string]struct{}{}
	for _, start := range starts {
		if start == "" {
			continue
		}
		clean := filepath.Clean(start)
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		if root, ok := findCLIUpgradeRuntimeRoot(clean); ok {
			return root, nil
		}
	}
	return "", fmt.Errorf("could not locate CLI upgrade runtime assets from the runectx executable location")
}

func cliUpgradeRuntimeStartPaths(deps cliUpgradeRuntimeDeps) []string {
	starts := make([]string, 0, 3)
	if exe, err := deps.executable(); err == nil {
		exeDir := filepath.Dir(exe)
		starts = append(starts, exeDir)
		starts = append(starts, filepath.Join(exeDir, ".."))
		starts = append(starts, filepath.Join(exeDir, "..", "share", "runecontext"))
	}
	return starts
}

func normalizeCLIUpgradeRuntimeDeps(deps cliUpgradeRuntimeDeps) cliUpgradeRuntimeDeps {
	if deps.executable == nil {
		deps.executable = os.Executable
	}
	return deps
}

func findCLIUpgradeRuntimeRoot(start string) (string, bool) {
	current := start
	if info, err := os.Stat(current); err == nil && !info.IsDir() {
		current = filepath.Dir(current)
	}
	for {
		if isCLIUpgradeRuntimeRoot(current) {
			return current, true
		}
		next := filepath.Dir(current)
		if next == current {
			return "", false
		}
		current = next
	}
}

func isCLIUpgradeRuntimeRoot(root string) bool {
	metadataPath := filepath.Join(root, "nix", "release", "metadata.nix")
	if _, err := os.Stat(metadataPath); err == nil {
		return true
	}
	shareRoot := filepath.Join(root, "share", "runecontext")
	if !isSchemaDir(filepath.Join(shareRoot, "schemas")) {
		return false
	}
	installerPath := filepath.Join(shareRoot, "installers", cliUpgradeInstallerScriptName())
	_, err := os.Stat(installerPath)
	return err == nil
}

func cliUpgradeInstallerScriptName() string {
	if runtime.GOOS == "windows" {
		return "install-runectx.ps1"
	}
	return "install-runectx.sh"
}

func repoRootForBundledReleaseAssets() (string, error) {
	executablePath, err := cliUpgradeExecutablePathFn()
	if err != nil {
		return "", err
	}
	resolvedExecutable, err := filepath.EvalSymlinks(executablePath)
	if err != nil {
		return "", fmt.Errorf("resolve runectx executable path: %w", err)
	}
	return findRepoRootForReleaseMetadata(filepath.Dir(resolvedExecutable))
}

func readLatestCLIReleaseFromRuntimeRoot(root string) (string, error) {
	manifestPath := filepath.Join(root, "release-manifest.json")
	if isCLIUpgradeInstalledShareRoot(root) {
		manifestPath = filepath.Join(root, "share", "runecontext", "release-manifest.json")
	}
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", fmt.Errorf("read release manifest: %w", err)
	}
	if err := json.Unmarshal(raw, new(map[string]any)); err != nil {
		return "", fmt.Errorf("parse release manifest: %w", err)
	}
	descriptor, err := releaseManifestDescriptorFromJSON(raw)
	if err != nil {
		return "", err
	}
	releaseValue, ok := descriptor["release"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("release manifest metadata_descriptor.release must be an object")
	}
	version, ok := releaseValue["version"].(string)
	if !ok || strings.TrimSpace(version) == "" {
		return "", fmt.Errorf("release manifest metadata_descriptor.release.version must be a non-empty string")
	}
	return version, nil
}

func isCLIUpgradeInstalledShareRoot(root string) bool {
	return isSchemaDir(filepath.Join(root, "share", "runecontext", "schemas"))
}

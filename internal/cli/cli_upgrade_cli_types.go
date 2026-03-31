package cli

import (
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
var cliUpgradeGetwdFn = os.Getwd

type staticLatestCLIReleaseResolver struct{}

func (staticLatestCLIReleaseResolver) ResolveLatestRelease(currentVersion string) (string, error) {
	root, err := cliUpgradeRuntimeRoot()
	if err != nil {
		return "", fmt.Errorf("resolve latest release metadata root: %w", err)
	}
	version, err := ReadReleaseMetadataVersion(root)
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
	return locateCLIUpgradeRuntimeRoot(cliUpgradeRuntimeDeps{getwd: cliUpgradeGetwdFn, executable: cliUpgradeExecutablePathFn})
}

type cliUpgradeRuntimeDeps struct {
	getwd      func() (string, error)
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
	return "", fmt.Errorf("could not locate CLI upgrade runtime assets from the current working directory or executable location")
}

func cliUpgradeRuntimeStartPaths(deps cliUpgradeRuntimeDeps) []string {
	starts := make([]string, 0, 4)
	if exe, err := deps.executable(); err == nil {
		exeDir := filepath.Dir(exe)
		starts = append(starts, exeDir)
		starts = append(starts, filepath.Join(exeDir, ".."))
		starts = append(starts, filepath.Join(exeDir, "..", "share", "runecontext"))
	}
	if wd, err := deps.getwd(); err == nil {
		starts = append(starts, wd)
	}
	return starts
}

func normalizeCLIUpgradeRuntimeDeps(deps cliUpgradeRuntimeDeps) cliUpgradeRuntimeDeps {
	if deps.getwd == nil {
		deps.getwd = os.Getwd
	}
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

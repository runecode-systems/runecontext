package cli

import (
	"fmt"
	"os"
	"path/filepath"
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
	repoRoot, err := repoRootForBundledReleaseAssets()
	if err != nil {
		return "", fmt.Errorf("resolve latest release metadata root: %w", err)
	}
	version, err := ReadReleaseMetadataVersion(repoRoot)
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

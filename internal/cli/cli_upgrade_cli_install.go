package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type defaultCLIUpgradeInstaller struct{}

func (defaultCLIUpgradeInstaller) Apply(plan cliUpgradePlan) (cliUpgradePlan, error) {
	if plan.Availability != cliUpgradeAvailabilityAvailable {
		plan.Changed = false
		return plan, nil
	}
	if err := installCLIReleaseVersion(plan.TargetRelease); err != nil {
		return plan, err
	}
	plan.Changed = true
	if binary, err := os.Executable(); err == nil {
		plan.UpdatedBinaryPath = binary
	}
	return plan, nil
}

func installCLIReleaseVersion(version string) error {
	script, args, err := installerCommandForCurrentPlatform(version)
	if err != nil {
		return err
	}
	cmd := exec.Command(script, args...)
	if output, runErr := cmd.CombinedOutput(); runErr != nil {
		return fmt.Errorf("run installer command: %w: %s", runErr, strings.TrimSpace(string(output)))
	}
	return nil
}

func installerCommandForCurrentPlatform(version string) (string, []string, error) {
	version = strings.TrimSpace(version)
	if !semverLikePattern.MatchString(version) {
		return "", nil, fmt.Errorf("target release %q must look like a semantic version", version)
	}
	runtimeRoot, err := cliUpgradeRuntimeRoot()
	if err != nil {
		return "", nil, err
	}
	if runtime.GOOS == "windows" {
		script := cliUpgradeInstallerPath(runtimeRoot, "install-runectx.ps1")
		if statErr := ensureTrustedInstallerScript(script); statErr != nil {
			return "", nil, fmt.Errorf("locate installer script: %w", statErr)
		}
		return "powershell", []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script, "-Version", "v" + version, "-Yes"}, nil
	}
	script := cliUpgradeInstallerPath(runtimeRoot, "install-runectx.sh")
	if statErr := ensureTrustedInstallerScript(script); statErr != nil {
		return "", nil, fmt.Errorf("locate installer script: %w", statErr)
	}
	return "bash", []string{script, "--version", "v" + version, "--yes"}, nil
}

func cliUpgradeInstallerPath(root, scriptName string) string {
	installed := filepath.Join(root, "share", "runecontext", "installers", scriptName)
	if _, err := os.Stat(installed); err == nil {
		return installed
	}
	return filepath.Join(root, "scripts", scriptName)
}

func ensureTrustedInstallerScript(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("installer script must not be a symlink: %s", filepath.ToSlash(path))
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("installer script must be a regular file: %s", filepath.ToSlash(path))
	}
	return nil
}

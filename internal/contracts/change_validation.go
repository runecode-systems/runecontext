package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validateChangeCommandInputs(v *Validator, loaded *LoadedProject) error {
	if v == nil {
		return fmt.Errorf("validator is required")
	}
	if loaded == nil {
		return fmt.Errorf("loaded project is required")
	}
	return nil
}

func validateWritableChangeCommand(v *Validator, loaded *LoadedProject) error {
	if err := validateChangeCommandInputs(v, loaded); err != nil {
		return err
	}
	return requireWritableChangeSource(loaded)
}

func validateChangeMutation(v *Validator, projectRoot string) error {
	validated, err := validateProjectAfterChangeMutation(v, projectRoot)
	if err != nil {
		return err
	}
	_ = validated.Close()
	return nil
}

func requireWritableChangeSource(loaded *LoadedProject) error {
	if loaded == nil || loaded.Resolution == nil {
		return fmt.Errorf("loaded project resolution is required")
	}
	switch loaded.Resolution.SourceMode {
	case SourceModeEmbedded, SourceModePath:
		return nil
	default:
		return fmt.Errorf("change write operations are only supported for embedded and local path sources in alpha.3")
	}
}

func writableContentRoot(loaded *LoadedProject) (string, error) {
	if err := requireWritableChangeSource(loaded); err != nil {
		return "", err
	}
	if loaded.Resolution.SourceMode == SourceModeEmbedded {
		return loaded.Resolution.MaterializedRoot(), nil
	}
	if filepath.IsAbs(loaded.Resolution.SourceRoot) {
		return filepath.Clean(loaded.Resolution.SourceRoot), nil
	}
	return filepath.Clean(filepath.Join(loaded.Resolution.ProjectRoot, loaded.Resolution.SourceRoot)), nil
}

func validateRequestedMode(mode ChangeMode) error {
	if mode == "" || mode == ChangeModeMinimum || mode == ChangeModeFull {
		return nil
	}
	return fmt.Errorf("change mode must be %q or %q", ChangeModeMinimum, ChangeModeFull)
}

func validateChangeTypeValue(changeType string) error {
	if strings.TrimSpace(changeType) == "" {
		return fmt.Errorf("change type is required")
	}
	if strings.HasPrefix(changeType, "x-") {
		return nil
	}
	if isBuiltInChangeType(changeType) {
		return nil
	}
	return fmt.Errorf("unsupported change type %q", changeType)
}

func isBuiltInChangeType(changeType string) bool {
	switch changeType {
	case "project", "feature", "bug", "standard", "chore":
		return true
	default:
		return false
	}
}

func validateCloseVerificationStatus(current, requested string) error {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		if current == "pending" {
			return fmt.Errorf("change close requires --verification-status when the current verification_status is pending")
		}
		return nil
	}
	if requested == "pending" {
		return fmt.Errorf("change close must not set verification_status to pending")
	}
	if isSupportedVerificationStatus(requested) {
		return nil
	}
	return fmt.Errorf("unsupported verification_status %q", requested)
}

func isSupportedVerificationStatus(status string) bool {
	switch status {
	case "passed", "failed", "skipped":
		return true
	default:
		return false
	}
}

func inferChangeMode(changeDir string) ChangeMode {
	for _, name := range []string{"design.md", "verification.md"} {
		if _, err := os.Stat(filepath.Join(changeDir, name)); err == nil {
			return ChangeModeFull
		}
	}
	return ChangeModeMinimum
}

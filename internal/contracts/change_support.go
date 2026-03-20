package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func verificationAssumption(projectRoot string) string {
	_, assumption := inferVerificationCommands(projectRoot)
	return assumption
}

func normalizedChangeTime(now time.Time) time.Time {
	now = now.UTC()
	if now.IsZero() {
		return time.Now().UTC()
	}
	return now
}

func changeDirCleanup(path string) func(*error) {
	return func(err *error) {
		if path == "" || err == nil || *err == nil {
			return
		}
		if cleanupErr := removeAllPath(path); cleanupErr != nil {
			*err = fmt.Errorf("%v; cleanup also failed and manual removal may be required: %v", *err, cleanupErr)
		}
	}
}

func disableChangeDirCleanup(*error) {}

func inferVerificationCommands(projectRoot string) ([]string, string) {
	if justfileHasTestTarget(projectRoot) {
		return []string{"just test"}, "Inferred `just test` from the repository's justfile test target."
	}
	if fileExists(filepath.Join(projectRoot, "go.mod")) {
		return []string{"go test ./..."}, "Inferred `go test ./...` from the repository's Go module layout."
	}
	if fileExists(filepath.Join(projectRoot, "package.json")) {
		return []string{"npm test"}, "Inferred `npm test` from the repository's package.json."
	}
	return nil, ""
}

func justfileHasTestTarget(projectRoot string) bool {
	data, err := os.ReadFile(filepath.Join(projectRoot, "justfile"))
	return err == nil && justfileTestTargetPattern.Match(data)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

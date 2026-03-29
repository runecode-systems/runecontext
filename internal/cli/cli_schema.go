package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

type schemaRootDeps struct {
	getwd      func() (string, error)
	executable func() (string, error)
}

func locateSchemaRoot() (string, error) {
	return locateSchemaRootWithDeps(schemaRootDeps{getwd: os.Getwd, executable: os.Executable})
}

func locateSchemaRootWithDeps(deps schemaRootDeps) (string, error) {
	starts := schemaRootStartPaths(normalizeSchemaRootDeps(deps))

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
		if root, ok := findSchemaRoot(clean); ok {
			return root, nil
		}
	}
	return "", fmt.Errorf("could not locate RuneContext schemas from the current working directory or executable location")
}

func normalizeSchemaRootDeps(deps schemaRootDeps) schemaRootDeps {
	if deps.getwd == nil {
		deps.getwd = os.Getwd
	}
	if deps.executable == nil {
		deps.executable = os.Executable
	}
	return deps
}

func schemaRootStartPaths(deps schemaRootDeps) []string {
	starts := make([]string, 0, 4)
	if wd, err := deps.getwd(); err == nil {
		starts = append(starts, wd)
	}
	if exe, err := deps.executable(); err == nil {
		exeDir := filepath.Dir(exe)
		starts = append(starts, exeDir)
		starts = append(starts, filepath.Join(exeDir, "..", "share", "runecontext"))
	}
	return starts
}

func findSchemaRoot(start string) (string, bool) {
	current := start
	if info, err := os.Stat(current); err == nil && !info.IsDir() {
		current = filepath.Dir(current)
	}
	for {
		if isSchemaDir(current) {
			return current, true
		}
		candidate := filepath.Join(current, "schemas")
		if isSchemaDir(candidate) {
			return candidate, true
		}
		next := filepath.Dir(current)
		if next == current {
			return "", false
		}
		current = next
	}
}

func isSchemaDir(path string) bool {
	for _, name := range []string{"runecontext.schema.json", "bundle.schema.json", "change-status.schema.json", "context-pack.schema.json", "assurance-baseline.schema.json", "assurance-receipt.schema.json", "assurance-imported-history.schema.json"} {
		if _, err := os.Stat(filepath.Join(path, name)); err != nil {
			return false
		}
	}
	return true
}

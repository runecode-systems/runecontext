package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

func locateSchemaRoot() (string, error) {
	starts := make([]string, 0, 2)
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}

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
	for _, name := range []string{"runecontext.schema.json", "bundle.schema.json", "change-status.schema.json", "context-pack.schema.json"} {
		if _, err := os.Stat(filepath.Join(path, name)); err != nil {
			return false
		}
	}
	return true
}

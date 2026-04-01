package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

func locateAdaptersRoot() (string, error) {
	schemaRoot, err := locateSchemaRoot()
	if err != nil {
		return "", err
	}
	projectRoot := filepath.Dir(schemaRoot)
	candidates := []string{
		filepath.Join(projectRoot, "build", "generated", "adapters"),
		filepath.Join(projectRoot, "adapters"),
	}
	for _, candidate := range candidates {
		if isAdapterPackRoot(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not locate installed adapter packs")
}

func isAdapterPackRoot(path string) bool {
	if !isDirectory(path) {
		return false
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "source" {
			continue
		}
		workflowPath := filepath.Join(path, entry.Name(), "workflow.json")
		if fileExists(workflowPath) {
			return true
		}
	}
	return false
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

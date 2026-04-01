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
		filepath.Join(schemaRoot, "adapters"),
		filepath.Join(projectRoot, "adapters"),
	}
	for _, candidate := range candidates {
		if isDirectory(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not locate installed adapter packs")
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

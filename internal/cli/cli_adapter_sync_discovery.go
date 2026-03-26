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
	candidates := []string{
		filepath.Join(schemaRoot, "adapters"),
		filepath.Join(filepath.Dir(schemaRoot), "adapters"),
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

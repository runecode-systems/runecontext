package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func rewriteRunecontextVersion(data []byte, target string) ([]byte, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("parse runecontext.yaml: %w", err)
	}
	if len(document.Content) == 0 {
		return nil, fmt.Errorf("runecontext.yaml is missing root document")
	}
	mapping := document.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("runecontext.yaml root must be a mapping")
	}
	if !setRunecontextVersion(mapping, target) {
		return nil, fmt.Errorf("runecontext.yaml is missing runecontext_version")
	}
	updated, err := yaml.Marshal(&document)
	if err != nil {
		return nil, fmt.Errorf("render runecontext.yaml: %w", err)
	}
	return updated, nil
}

func setRunecontextVersion(mapping *yaml.Node, target string) bool {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key := mapping.Content[i]
		if key.Kind != yaml.ScalarNode || key.Value != "runecontext_version" {
			continue
		}
		value := mapping.Content[i+1]
		value.Kind = yaml.ScalarNode
		value.Tag = "!!str"
		value.Value = target
		return true
	}
	return false
}

func configFileMode(path string) os.FileMode {
	info, err := os.Stat(path)
	if err != nil {
		return 0o644
	}
	return info.Mode().Perm()
}

func writeAtomicUpgradeConfig(path string, data []byte, mode os.FileMode) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".upgrade-config-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tempPath)
		}
	}()
	if err := temp.Chmod(mode); err != nil {
		_ = temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

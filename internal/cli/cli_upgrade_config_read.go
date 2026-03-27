package cli

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func readRunecontextVersionFromConfig(configPath string) string {
	rootMap, ok := readRootConfigMap(configPath)
	if !ok {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(rootMap["runecontext_version"]))
}

func readSourcePathFromConfig(configPath string) string {
	rootMap, ok := readRootConfigMap(configPath)
	if !ok {
		return ""
	}
	sourceMap, ok := rootMap["source"].(map[string]any)
	if !ok {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(sourceMap["path"]))
}

func readRootConfigMap(configPath string) (map[string]any, bool) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, false
	}
	rootMap := map[string]any{}
	if err := yaml.Unmarshal(data, &rootMap); err != nil {
		return nil, false
	}
	return rootMap, true
}

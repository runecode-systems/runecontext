package cli

import (
	"fmt"
)

func generateCompletionScript(registry MetadataRegistry, shell string) (string, error) {
	index := buildCompletionIndex(registry)
	switch shell {
	case "bash":
		return buildBashCompletionScript(index), nil
	case "zsh":
		return buildZshCompletionScript(index), nil
	case "fish":
		return buildFishCompletionScript(index), nil
	default:
		return "", fmt.Errorf("unsupported shell %q; expected one of: bash, zsh, fish", shell)
	}
}

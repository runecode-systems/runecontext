package cli

import (
	"fmt"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

const (
	standardListCommand   = "standard_list"
	standardCreateCommand = "standard_create"
	standardUpdateCommand = "standard_update"
)

type standardListRequest struct {
	root         string
	explicitRoot bool
	scopePaths   []string
	focus        string
	statuses     []string
}

type standardCreateRequest struct {
	root                    string
	explicitRoot            bool
	path                    string
	id                      string
	title                   string
	status                  string
	replacedBy              string
	aliases                 []string
	suggestedContextBundles []string
	body                    string
}

type standardUpdateRequest struct {
	root                           string
	explicitRoot                   bool
	path                           string
	title                          string
	status                         string
	replacedBy                     string
	clearReplacedBy                bool
	replaceAliases                 bool
	aliases                        []string
	replaceSuggestedContextBundles bool
	suggestedContextBundles        []string
}

func parseStandardStatuses(items []string) []contracts.StandardStatus {
	out := make([]contracts.StandardStatus, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		out = append(out, contracts.StandardStatus(trimmed))
	}
	return out
}

func buildStandardMutationOutput(absRoot string, loaded *contracts.LoadedProject, command string, result *contracts.StandardMutationResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", command},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"standard_path", result.Path},
		{"mutation_performed", boolString(len(result.ChangedFiles) > 0)},
	}
	return appendChangedFiles(output, result.ChangedFiles)
}

func appendStandardMutationExplainLines(lines []line, command string, result *contracts.StandardMutationResult) []line {
	if result == nil {
		return lines
	}
	return append(lines,
		line{"explain_scope", "standards-authoring"},
		line{"explain_operation", command},
		line{"explain_mutation_count_reason", fmt.Sprintf("%d standard artifact files changed through validated transactional writes", len(result.ChangedFiles))},
	)
}

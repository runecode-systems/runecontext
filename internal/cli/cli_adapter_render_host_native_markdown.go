package cli

import (
	"fmt"
	"slices"
	"strings"
)

func renderHostNativeOperationMarkdown(request adapterRenderRequest) (string, error) {
	request.role = normalizeHostNativeRole(request.role)
	if request.operation == "index" {
		return renderHostNativeIndexMarkdown(request)
	}
	flow, ok := flowByOperation(request.tool, request.operation)
	if !ok {
		return "", fmt.Errorf("adapter render-host-native operation %q not defined for tool %q", request.operation, request.tool)
	}
	meta, err := hostNativeRenderMetadata(request, flow)
	if err != nil {
		return "", err
	}
	lines := []string{
		"- canonical_flow_source: `" + meta.source + "`",
		"- adapter_role: `" + meta.role + "`",
		"- operation_identifier: `" + meta.operationID + "`",
		"- command_path: `" + meta.commandPath + "`",
		"- usage: `" + meta.usage + "`",
	}
	if len(meta.requiredFlags) > 0 {
		lines = append(lines, "- required_flags: `"+strings.Join(meta.requiredFlags, " ")+"`")
	}
	if len(meta.requiredPositionals) > 0 {
		lines = append(lines, "- required_positionals: `"+strings.Join(meta.requiredPositionals, " ")+"`")
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func renderHostNativeIndexMarkdown(request adapterRenderRequest) (string, error) {
	if request.role != hostNativeKindDiscoverabilityShim {
		return "", fmt.Errorf("adapter render-host-native index requires role %q", hostNativeKindDiscoverabilityShim)
	}
	flows := toolFlowMappings(request.tool)
	if len(flows) == 0 {
		return "", fmt.Errorf("adapter render-host-native index is not defined for tool %q", request.tool)
	}
	lines := []string{
		"- canonical_flow_source: `adapters/" + request.tool + "/flows/*.md`",
		"- adapter_role: `" + hostNativeKindDiscoverabilityShim + "`",
		"- operation_identifier: `runecontext:index`",
	}
	for _, flow := range flows {
		commandPath := strings.ReplaceAll(flow.id, "-", " ")
		command, ok := commandByPath(commandPath)
		if !ok {
			continue
		}
		lines = append(lines,
			"- operation: `runecontext:"+flow.id+"`",
			"  - command_path: `"+commandPath+"`",
			"  - usage: `"+command.Usage+"`",
		)
	}
	return strings.Join(lines, "\n") + "\n", nil
}

type hostNativeRenderMeta struct {
	operationID         string
	source              string
	role                string
	commandPath         string
	usage               string
	requiredFlags       []string
	requiredPositionals []string
}

func hostNativeRenderMetadata(request adapterRenderRequest, flow hostNativeFlow) (hostNativeRenderMeta, error) {
	path := strings.ReplaceAll(flow.id, "-", " ")
	command, ok := commandByPath(path)
	if !ok {
		return hostNativeRenderMeta{}, fmt.Errorf("adapter render-host-native metadata missing command path %q", path)
	}
	role := hostNativeKindFlowAsset
	if request.role == hostNativeKindDiscoverabilityShim {
		role = hostNativeKindDiscoverabilityShim
	}
	return hostNativeRenderMeta{
		operationID:         "runecontext:" + flow.id,
		source:              flow.source,
		role:                role,
		commandPath:         path,
		usage:               command.Usage,
		requiredFlags:       requiredFlagsFromUsage(command.Usage, command.Flags),
		requiredPositionals: requiredPositionalsFromUsage(command.Usage, command.Positionals),
	}, nil
}

func commandByPath(path string) (CommandMetadata, bool) {
	registry := CommandMetadataRegistry()
	var walk func(items []CommandMetadata) (CommandMetadata, bool)
	walk = func(items []CommandMetadata) (CommandMetadata, bool) {
		for _, item := range items {
			if item.Path == path {
				return item, true
			}
			if found, ok := walk(item.Subcommands); ok {
				return found, true
			}
		}
		return CommandMetadata{}, false
	}
	return walk(registry.Commands)
}

func requiredFlagsFromUsage(usage string, flags []FlagMetadata) []string {
	required := make([]string, 0)
	for _, flag := range flags {
		if !strings.Contains(usage, "["+flag.Name) {
			required = append(required, flag.Name)
		}
	}
	slices.Sort(required)
	return required
}

func requiredPositionalsFromUsage(usage string, positionals []PositionalMetadata) []string {
	required := make([]string, 0)
	for _, positional := range positionals {
		if positional.Optional || positional.Variadic {
			continue
		}
		token := " " + positional.Name
		if strings.Contains(usage, token) && !strings.Contains(usage, "["+positional.Name+"]") {
			required = append(required, positional.Name)
		}
	}
	slices.Sort(required)
	return required
}

func supportsShellInjection(tool string) bool {
	return tool == "opencode" || tool == "claude-code"
}

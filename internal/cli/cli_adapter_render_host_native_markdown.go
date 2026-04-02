package cli

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

func renderHostNativeOperationMarkdown(request adapterRenderRequest) (string, error) {
	request.role = normalizeHostNativeRole(request.role)
	if request.operation == "index" {
		return renderHostNativeIndexMarkdown(request)
	}
	flow, ok, err := flowByOperation(request.tool, request.operation)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("adapter render-host-native operation %q not defined for tool %q", request.operation, request.tool)
	}
	meta, err := hostNativeRenderMetadata(request, flow)
	if err != nil {
		return "", err
	}
	lines := []string{
		"- canonical_flow_source: `" + meta.source + "`",
		"- canonical_workflow_contract: `" + meta.contractSource + "`",
		"- adapter_role: `" + meta.role + "`",
		"- operation_identifier: `" + meta.operationID + "`",
		"- command_path: `" + meta.commandPath + "`",
		"- usage: `" + meta.usage + "`",
		"- required_outcome: " + meta.requiredOutcome,
		"- interaction_rule: " + hostNativeNoQuestionRule,
	}
	lines = appendHostNativeStructuredSections(lines, meta)
	if len(meta.requiredFlags) > 0 {
		lines = append(lines, "- required_flags: `"+strings.Join(meta.requiredFlags, " ")+"`")
	}
	if len(meta.requiredPositionals) > 0 {
		lines = append(lines, "- required_positionals: `"+strings.Join(meta.requiredPositionals, " ")+"`")
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func appendHostNativeStructuredSections(lines []string, meta hostNativeRenderMeta) []string {
	lines = appendNestedList(lines, "guardrails", meta.guardrails)
	lines = appendNestedList(lines, "inputs_to_gather", meta.inputsToGather)
	lines = appendNestedList(lines, "decision_rules", meta.decisionRules)
	lines = appendNumberedNestedList(lines, "workflow_steps", meta.workflowSteps)
	if meta.stopCondition != "" {
		lines = append(lines, "- stop_condition: "+meta.stopCondition)
	}
	lines = appendBacktickNestedList(lines, "recommended_next_commands", meta.recommendedNextCommands)
	return appendExampleList(lines, meta.examples)
}

func appendNestedList(lines []string, key string, items []string) []string {
	if len(items) == 0 {
		return lines
	}
	lines = append(lines, "- "+key+":")
	for _, item := range items {
		lines = append(lines, "  - "+item)
	}
	return lines
}

func appendNumberedNestedList(lines []string, key string, items []string) []string {
	if len(items) == 0 {
		return lines
	}
	lines = append(lines, "- "+key+":")
	for index, item := range items {
		lines = append(lines, fmt.Sprintf("  %d. %s", index+1, item))
	}
	return lines
}

func appendBacktickNestedList(lines []string, key string, items []string) []string {
	if len(items) == 0 {
		return lines
	}
	lines = append(lines, "- "+key+":")
	for _, item := range items {
		lines = append(lines, "  - `"+item+"`")
	}
	return lines
}

func appendExampleList(lines []string, examples []adapterWorkflowExample) []string {
	if len(examples) == 0 {
		return lines
	}
	lines = append(lines, "- examples:")
	for _, example := range examples {
		lines = append(lines, "  - scenario: "+example.Scenario, "    user_prompt: |")
		lines = appendIndentedMultiline(lines, "      ", example.UserPrompt)
		lines = append(lines, "    assistant_response: |")
		lines = appendIndentedMultiline(lines, "      ", example.AssistantResponse)
	}
	return lines
}

func appendIndentedMultiline(lines []string, indent string, value string) []string {
	for _, part := range strings.Split(value, "\n") {
		lines = append(lines, indent+part)
	}
	return lines
}

func renderHostNativeIndexMarkdown(request adapterRenderRequest) (string, error) {
	if request.role != hostNativeKindDiscoverabilityShim {
		return "", fmt.Errorf("adapter render-host-native index requires role %q", hostNativeKindDiscoverabilityShim)
	}
	flows, err := toolFlowMappings(request.tool)
	if err != nil {
		return "", err
	}
	if len(flows) == 0 {
		return "", fmt.Errorf("adapter render-host-native index is not defined for tool %q", request.tool)
	}
	referenceRoot, err := adapterReferenceRoot()
	if err != nil {
		return "", err
	}
	lines := []string{
		"- canonical_flow_source: `" + filepath.ToSlash(filepath.Join(referenceRoot, request.tool, "flows", "*.md")) + "`",
		"- canonical_workflow_contract: `" + workflowContractReferencePath(referenceRoot, request.tool) + "`",
		"- adapter_role: `" + hostNativeKindDiscoverabilityShim + "`",
		"- operation_identifier: `runecontext:index`",
		"- interaction_rule: " + hostNativeNoQuestionRule,
	}
	for _, flow := range flows {
		commandPath := flow.commandPath
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
	operationID             string
	source                  string
	contractSource          string
	role                    string
	commandPath             string
	usage                   string
	requiredOutcome         string
	guardrails              []string
	inputsToGather          []string
	decisionRules           []string
	workflowSteps           []string
	stopCondition           string
	recommendedNextCommands []string
	examples                []adapterWorkflowExample
	requiredFlags           []string
	requiredPositionals     []string
}

func hostNativeRenderMetadata(request adapterRenderRequest, flow hostNativeFlow) (hostNativeRenderMeta, error) {
	path := flow.commandPath
	command, ok := commandByPath(path)
	if !ok {
		return hostNativeRenderMeta{}, fmt.Errorf("adapter render-host-native metadata missing command path %q", path)
	}
	role := hostNativeKindFlowAsset
	if request.role == hostNativeKindDiscoverabilityShim {
		role = hostNativeKindDiscoverabilityShim
	}
	referenceRoot, err := adapterReferenceRoot()
	if err != nil {
		return hostNativeRenderMeta{}, err
	}
	return hostNativeRenderMeta{
		operationID:             "runecontext:" + flow.id,
		source:                  flow.source,
		contractSource:          workflowContractReferencePath(referenceRoot, request.tool),
		role:                    role,
		commandPath:             path,
		usage:                   command.Usage,
		requiredOutcome:         flow.requiredOutcome,
		guardrails:              append([]string{}, flow.guardrails...),
		inputsToGather:          append([]string{}, flow.inputsToGather...),
		decisionRules:           append([]string{}, flow.decisionRules...),
		workflowSteps:           append([]string{}, flow.workflowSteps...),
		stopCondition:           flow.stopCondition,
		recommendedNextCommands: append([]string{}, flow.recommendedNextCommands...),
		examples:                append([]adapterWorkflowExample{}, flow.examples...),
		requiredFlags:           requiredFlagsFromMetadata(command.Flags),
		requiredPositionals:     requiredPositionalsFromUsage(command.Usage, command.Positionals),
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

func requiredFlagsFromMetadata(flags []FlagMetadata) []string {
	required := make([]string, 0)
	for _, flag := range flags {
		if flag.Required {
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

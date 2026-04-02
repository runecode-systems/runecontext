package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeFlows(targetDir string, tool toolDefinition, flows []flowDefinition) error {
	flowsDir := filepath.Join(targetDir, generatedFlowsDir)
	if err := os.MkdirAll(flowsDir, 0o755); err != nil {
		return fmt.Errorf("create flows dir: %w", err)
	}
	for _, flow := range flows {
		if err := writeFlowFile(flowsDir, tool, flow); err != nil {
			return err
		}
	}
	return nil
}

func writeFlowFile(flowsDir string, tool toolDefinition, flow flowDefinition) error {
	path := filepath.Join(flowsDir, flow.ID+".md")
	lines := buildFlowHeaderLines(tool, flow)
	lines = appendFlowStructuredSections(lines, flow)
	lines = append(lines,
		"",
		"## Intent",
		"",
		"- "+flow.Description,
		"",
		"## Command Mapping",
		"",
		"```sh",
		flow.Usage,
		"```",
		"",
		"## Examples",
	)
	lines = appendFlowExampleSections(lines, flow.Examples)
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		return fmt.Errorf("write flow %q: %w", flow.ID, err)
	}
	return nil
}

func buildFlowHeaderLines(tool toolDefinition, flow flowDefinition) []string {
	return []string{
		"# " + tool.DisplayName + " Flow: " + flow.CommandPath,
		"",
		tool.FlowIntro,
		"",
	}
}

func appendFlowStructuredSections(lines []string, flow flowDefinition) []string {
	lines = append(lines,
		"## Required Outcome",
		"",
		"- "+flow.RequiredOutcome,
		"",
		"## Guardrails",
		"",
	)
	lines = appendBulletLines(lines, flow.Guardrails)
	lines = append(lines,
		"",
		"## Inputs To Gather",
		"",
	)
	lines = appendBulletLines(lines, flow.InputsToGather)
	lines = append(lines,
		"",
		"## Decision Rules",
		"",
	)
	lines = appendBulletLines(lines, flow.DecisionRules)
	lines = append(lines,
		"",
		"## Workflow Steps",
		"",
	)
	lines = appendNumberedLines(lines, flow.WorkflowSteps)
	lines = append(lines,
		"",
		"## Stop Condition",
		"",
		"- "+flow.StopCondition,
		"",
		"## Recommended Next Commands",
		"",
	)
	return appendBulletLines(lines, flow.RecommendedNextCommands)
}

func appendFlowExampleSections(lines []string, examples []flowExample) []string {
	for _, example := range examples {
		lines = append(lines,
			"",
			"### "+example.Scenario,
			"",
			"User prompt:",
			"",
			"```text",
			example.UserPrompt,
			"```",
			"",
			"Assistant response:",
			"",
			"```text",
			example.AssistantResponse,
			"```",
		)
	}
	return lines
}

func appendBulletLines(lines []string, values []string) []string {
	for _, value := range values {
		lines = append(lines, "- "+value)
	}
	return lines
}

func appendNumberedLines(lines []string, values []string) []string {
	for i, value := range values {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, value))
	}
	return lines
}

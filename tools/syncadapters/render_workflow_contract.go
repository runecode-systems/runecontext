package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type generatedWorkflowContract struct {
	SchemaVersion string                  `json:"schema_version"`
	Adapter       string                  `json:"adapter"`
	DisplayName   string                  `json:"display_name"`
	FlowIntro     string                  `json:"flow_intro"`
	Flows         []generatedWorkflowFlow `json:"flows"`
}

type generatedWorkflowFlow struct {
	ID                      string        `json:"id"`
	CommandPath             string        `json:"command_path"`
	Description             string        `json:"description"`
	Usage                   string        `json:"usage"`
	RequiredOutcome         string        `json:"required_outcome"`
	Guardrails              []string      `json:"guardrails"`
	InputsToGather          []string      `json:"inputs_to_gather"`
	DecisionRules           []string      `json:"decision_rules"`
	WorkflowSteps           []string      `json:"workflow_steps"`
	StopCondition           string        `json:"stop_condition"`
	RecommendedNextCommands []string      `json:"recommended_next_commands"`
	Examples                []flowExample `json:"examples"`
}

func writeWorkflowContract(targetDir, toolID string, tool toolDefinition, flows []flowDefinition) error {
	path := filepath.Join(targetDir, "workflow.json")
	contract := generatedWorkflowContract{
		SchemaVersion: "1",
		Adapter:       toolID,
		DisplayName:   tool.DisplayName,
		FlowIntro:     tool.FlowIntro,
		Flows:         make([]generatedWorkflowFlow, 0, len(flows)),
	}
	for _, flow := range flows {
		contract.Flows = append(contract.Flows, generatedWorkflowFlow{
			ID:                      flow.ID,
			CommandPath:             flow.CommandPath,
			Description:             flow.Description,
			Usage:                   flow.Usage,
			RequiredOutcome:         flow.RequiredOutcome,
			Guardrails:              append([]string{}, flow.Guardrails...),
			InputsToGather:          append([]string{}, flow.InputsToGather...),
			DecisionRules:           append([]string{}, flow.DecisionRules...),
			WorkflowSteps:           append([]string{}, flow.WorkflowSteps...),
			StopCondition:           flow.StopCondition,
			RecommendedNextCommands: append([]string{}, flow.RecommendedNextCommands...),
			Examples:                append([]flowExample{}, flow.Examples...),
		})
	}
	raw, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		return fmt.Errorf("encode workflow contract: %w", err)
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type flowDefinition struct {
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

type flowExample struct {
	Scenario          string `json:"scenario"`
	UserPrompt        string `json:"user_prompt"`
	AssistantResponse string `json:"assistant_response"`
}

type toolDefinition struct {
	DisplayName  string            `json:"display_name"`
	FlowIntro    string            `json:"flow_intro"`
	Capabilities map[string]string `json:"capabilities"`
}

func loadFlowDefinitions(root string) ([]flowDefinition, error) {
	pattern := filepath.Join(root, "adapters", "source", "shared", "flows", "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("list flow definitions: %w", err)
	}
	if len(files) == 0 {
		return nil, errors.New("no adapter source flow definitions found")
	}
	defs := make([]flowDefinition, 0, len(files))
	seen := make(map[string]struct{}, len(files))
	for _, path := range files {
		var def flowDefinition
		if err := loadJSON(path, &def); err != nil {
			return nil, err
		}
		if err := validateFlowDefinition(def); err != nil {
			return nil, fmt.Errorf("invalid flow definition %q: %w", path, err)
		}
		if _, ok := seen[def.ID]; ok {
			return nil, fmt.Errorf("duplicate flow definition id %q", def.ID)
		}
		seen[def.ID] = struct{}{}
		defs = append(defs, def)
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].ID < defs[j].ID })
	return defs, nil
}

func validateFlowDefinition(def flowDefinition) error {
	if strings.TrimSpace(def.ID) == "" {
		return errors.New("id is required")
	}
	if err := validateFlowRequiredStrings(def); err != nil {
		return err
	}
	if err := validateFlowLists(def); err != nil {
		return err
	}
	return validateFlowExamples(def.Examples)
}

func validateFlowRequiredStrings(def flowDefinition) error {
	checks := []struct {
		name  string
		value string
	}{
		{name: "command_path", value: def.CommandPath},
		{name: "description", value: def.Description},
		{name: "usage", value: def.Usage},
		{name: "required_outcome", value: def.RequiredOutcome},
		{name: "stop_condition", value: def.StopCondition},
	}
	for _, check := range checks {
		if strings.TrimSpace(check.value) == "" {
			return fmt.Errorf("%s is required", check.name)
		}
	}
	return nil
}

func validateFlowLists(def flowDefinition) error {
	listChecks := []struct {
		name   string
		values []string
	}{
		{name: "guardrails", values: def.Guardrails},
		{name: "inputs_to_gather", values: def.InputsToGather},
		{name: "decision_rules", values: def.DecisionRules},
		{name: "workflow_steps", values: def.WorkflowSteps},
		{name: "recommended_next_commands", values: def.RecommendedNextCommands},
	}
	for _, check := range listChecks {
		if err := validateNonEmptyList(check.name, check.values); err != nil {
			return err
		}
	}
	return nil
}

func validateFlowExamples(examples []flowExample) error {
	if len(examples) == 0 {
		return errors.New("examples are required")
	}
	for i := range examples {
		if err := validateFlowExample(examples[i], i); err != nil {
			return err
		}
	}
	return nil
}

func validateNonEmptyList(name string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("%s is required", name)
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s must not contain empty entries", name)
		}
	}
	return nil
}

func validateFlowExample(example flowExample, index int) error {
	prefix := fmt.Sprintf("examples[%d]", index)
	if strings.TrimSpace(example.Scenario) == "" {
		return fmt.Errorf("%s.scenario is required", prefix)
	}
	if strings.TrimSpace(example.UserPrompt) == "" {
		return fmt.Errorf("%s.user_prompt is required", prefix)
	}
	if strings.TrimSpace(example.AssistantResponse) == "" {
		return fmt.Errorf("%s.assistant_response is required", prefix)
	}
	return nil
}

func loadToolDefinitions(root string) (map[string]toolDefinition, error) {
	pattern := filepath.Join(root, "adapters", "source", "tools", "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("list tool definitions: %w", err)
	}
	if len(files) == 0 {
		return nil, errors.New("no adapter source tool definitions found")
	}
	tools := make(map[string]toolDefinition, len(files))
	for _, path := range files {
		toolID := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		var def toolDefinition
		if err := loadJSON(path, &def); err != nil {
			return nil, err
		}
		if err := validateToolDefinition(toolID, def); err != nil {
			return nil, err
		}
		tools[toolID] = def
	}
	return tools, nil
}

func validateToolDefinition(toolID string, def toolDefinition) error {
	if strings.TrimSpace(toolID) == "" {
		return errors.New("tool id must be non-empty")
	}
	if strings.TrimSpace(def.DisplayName) == "" {
		return fmt.Errorf("tool %q display_name is required", toolID)
	}
	if strings.TrimSpace(def.FlowIntro) == "" {
		return fmt.Errorf("tool %q flow_intro is required", toolID)
	}
	for _, key := range capabilityOrder() {
		value := strings.TrimSpace(def.Capabilities[key])
		if value != "supported" && value != "optional" {
			return fmt.Errorf("tool %q capability %q must be supported or optional", toolID, key)
		}
	}
	return nil
}

func loadJSON(path string, target any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %q: %w", path, err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode %q: %w", path, err)
	}
	return nil
}

func sortedToolIDs(items map[string]toolDefinition) []string {
	ids := make([]string, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func capabilityOrder() []string {
	return []string{"prompts", "shell_access", "hooks", "dynamic_suggestions", "structured_output"}
}

func capabilityFallback(name, value string) string {
	if value == "supported" {
		switch name {
		case "prompts":
			return "Render static guidance and explicit runectx command proposals."
		case "shell_access":
			return "Show command steps and candidate data without execution."
		case "hooks":
			return "Run explicit runectx validate at review checkpoints."
		case "dynamic_suggestions":
			return "Use runectx completion suggest output as numbered choices."
		case "structured_output":
			return "Emit plain text summary with equivalent CLI flags/values."
		}
	}
	return "Use explicit runectx command proposals and manual review checkpoints."
}

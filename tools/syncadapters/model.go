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
	ID          string `json:"id"`
	CommandPath string `json:"command_path"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
}

type toolDefinition struct {
	DisplayName  string            `json:"display_name"`
	FlowIntro    string            `json:"flow_intro"`
	Capabilities map[string]string `json:"capabilities"`
}

func loadFlowDefinitions(root string) ([]flowDefinition, error) {
	path := filepath.Join(root, "adapters", "source", "shared", "flows.json")
	var defs []flowDefinition
	if err := loadJSON(path, &defs); err != nil {
		return nil, err
	}
	for i := range defs {
		if err := validateFlowDefinition(defs[i]); err != nil {
			return nil, fmt.Errorf("invalid flow definition %q: %w", defs[i].ID, err)
		}
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].ID < defs[j].ID })
	return defs, nil
}

func validateFlowDefinition(def flowDefinition) error {
	if strings.TrimSpace(def.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(def.CommandPath) == "" {
		return errors.New("command_path is required")
	}
	if strings.TrimSpace(def.Description) == "" {
		return errors.New("description is required")
	}
	if strings.TrimSpace(def.Usage) == "" {
		return errors.New("usage is required")
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

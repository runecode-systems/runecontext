package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type adapterWorkflowDocument struct {
	SchemaVersion string                `json:"schema_version"`
	Adapter       string                `json:"adapter"`
	DisplayName   string                `json:"display_name"`
	FlowIntro     string                `json:"flow_intro"`
	Flows         []adapterWorkflowFlow `json:"flows"`
}

type adapterWorkflowFlow struct {
	ID                      string                   `json:"id"`
	CommandPath             string                   `json:"command_path"`
	Description             string                   `json:"description"`
	Usage                   string                   `json:"usage"`
	RequiredOutcome         string                   `json:"required_outcome"`
	Guardrails              []string                 `json:"guardrails"`
	InputsToGather          []string                 `json:"inputs_to_gather"`
	DecisionRules           []string                 `json:"decision_rules"`
	WorkflowSteps           []string                 `json:"workflow_steps"`
	StopCondition           string                   `json:"stop_condition"`
	RecommendedNextCommands []string                 `json:"recommended_next_commands"`
	Examples                []adapterWorkflowExample `json:"examples"`
}

type adapterWorkflowExample struct {
	Scenario          string `json:"scenario"`
	UserPrompt        string `json:"user_prompt"`
	AssistantResponse string `json:"assistant_response"`
}

func toolFlowMappings(tool string) ([]hostNativeFlow, error) {
	doc, err := loadWorkflowDocument(tool)
	if err != nil {
		return nil, err
	}
	flows := make([]hostNativeFlow, 0, len(doc.Flows))
	for _, flow := range doc.Flows {
		flows = append(flows, hostNativeFlow{
			id:                      flow.ID,
			name:                    flow.CommandPath,
			description:             flow.Description,
			source:                  workflowMarkdownSource(tool, flow.ID),
			commandPath:             flow.CommandPath,
			usage:                   flow.Usage,
			requiredOutcome:         flow.RequiredOutcome,
			guardrails:              append([]string{}, flow.Guardrails...),
			inputsToGather:          append([]string{}, flow.InputsToGather...),
			decisionRules:           append([]string{}, flow.DecisionRules...),
			workflowSteps:           append([]string{}, flow.WorkflowSteps...),
			stopCondition:           flow.StopCondition,
			recommendedNextCommands: append([]string{}, flow.RecommendedNextCommands...),
			examples:                append([]adapterWorkflowExample{}, flow.Examples...),
		})
	}
	sort.Slice(flows, func(i, j int) bool { return flows[i].id < flows[j].id })
	return flows, nil
}

func loadWorkflowDocument(tool string) (adapterWorkflowDocument, error) {
	raw, path, err := loadWorkflowContractBytes(tool)
	if err != nil {
		return adapterWorkflowDocument{}, err
	}
	var doc adapterWorkflowDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return adapterWorkflowDocument{}, fmt.Errorf("decode adapter workflow contract %q: %w", path, err)
	}
	if err := validateWorkflowDocument(tool, doc); err != nil {
		return adapterWorkflowDocument{}, fmt.Errorf("invalid adapter workflow contract %q: %w", path, err)
	}
	return doc, nil
}

func loadWorkflowContractBytes(tool string) ([]byte, string, error) {
	path, err := workflowContractPath(tool)
	if err != nil {
		if genErr := ensureGeneratedAdapterPack(tool); genErr != nil {
			return nil, "", fmt.Errorf("%w (while recovering from %v)", genErr, err)
		}
		path, err = workflowContractPath(tool)
		if err != nil {
			return nil, "", err
		}
	}
	raw, err := os.ReadFile(path)
	if err == nil {
		return raw, path, nil
	}
	if !os.IsNotExist(err) {
		return nil, "", fmt.Errorf("read adapter workflow contract %q: %w", path, err)
	}
	if genErr := ensureGeneratedAdapterPack(tool); genErr != nil {
		return nil, "", fmt.Errorf("read adapter workflow contract %q: %w (regeneration failed: %v)", path, err, genErr)
	}
	raw, err = os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read adapter workflow contract %q: %w", path, err)
	}
	return raw, path, nil
}

func workflowContractPath(tool string) (string, error) {
	adaptersRoot, err := locateAdaptersRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(adaptersRoot, tool, "workflow.json"), nil
}

func ensureGeneratedAdapterPack(tool string) error {
	tool = strings.TrimSpace(tool)
	if tool == "" {
		return fmt.Errorf("adapter tool is required")
	}
	schemaRoot, err := locateSchemaRoot()
	if err != nil {
		return err
	}
	projectRoot := filepath.Dir(schemaRoot)
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err != nil {
		return fmt.Errorf("could not locate installed adapter packs")
	}
	cmd := exec.Command("go", "run", "./tools/syncadapters", "--root", projectRoot, "--output", "build/generated/adapters", "--tool", tool)
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return fmt.Errorf("generate adapter pack %q: %w", tool, err)
		}
		return fmt.Errorf("generate adapter pack %q: %w: %s", tool, err, message)
	}
	return nil
}

func workflowMarkdownSource(tool, flowID string) string {
	return "build/generated/adapters/" + tool + "/flows/" + flowID + ".md"
}

func validateWorkflowDocument(tool string, doc adapterWorkflowDocument) error {
	if doc.SchemaVersion == "" {
		return fmt.Errorf("schema_version is required")
	}
	if doc.Adapter != tool {
		return fmt.Errorf("adapter mismatch: expected %q, got %q", tool, doc.Adapter)
	}
	if len(doc.Flows) == 0 {
		return fmt.Errorf("flows are required")
	}
	seen := make(map[string]struct{}, len(doc.Flows))
	for _, flow := range doc.Flows {
		if flow.ID == "" {
			return fmt.Errorf("flow id is required")
		}
		if _, ok := seen[flow.ID]; ok {
			return fmt.Errorf("duplicate flow id %q", flow.ID)
		}
		seen[flow.ID] = struct{}{}
	}
	return nil
}

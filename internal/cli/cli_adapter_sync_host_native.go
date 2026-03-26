package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

const hostNativeOwnershipMarker = "runecontext-managed-artifact: host-native-v1"

type hostNativeArtifact struct {
	relPath string
	tool    string
	kind    string
	id      string
	shim    bool
	content []byte
}

type hostNativeFlow struct {
	id          string
	name        string
	description string
	source      string
}

func buildHostNativeArtifacts(tool string) ([]hostNativeArtifact, error) {
	switch tool {
	case "opencode":
		return buildOpenCodeHostNativeArtifacts(), nil
	case "claude-code":
		return buildClaudeCodeHostNativeArtifacts(), nil
	case "codex":
		return buildCodexHostNativeArtifacts(), nil
	case "generic":
		return nil, nil
	default:
		return nil, fmt.Errorf("adapter %q not found in installed adapter packs", tool)
	}
}

func buildOpenCodeHostNativeArtifacts() []hostNativeArtifact {
	flows := toolFlowMappings("opencode")
	artifacts := make([]hostNativeArtifact, 0, len(flows)*2)
	for _, flow := range flows {
		name := "runecontext-" + flow.id + ".md"
		artifacts = append(artifacts,
			hostNativeArtifact{
				relPath: filepath.ToSlash(filepath.Join(".opencode", "skills", name)),
				tool:    "opencode",
				kind:    "flow_asset",
				id:      "runecontext:" + flow.id,
				shim:    false,
				content: buildHostNativeFlowAssetContent("opencode", flow),
			},
			hostNativeArtifact{
				relPath: filepath.ToSlash(filepath.Join(".opencode", "commands", name)),
				tool:    "opencode",
				kind:    "discoverability_shim",
				id:      "runecontext:" + flow.id,
				shim:    true,
				content: buildHostNativeCommandShimContent("opencode", flow),
			},
		)
	}
	sortHostNativeArtifacts(artifacts)
	return artifacts
}

func buildClaudeCodeHostNativeArtifacts() []hostNativeArtifact {
	flows := toolFlowMappings("claude-code")
	artifacts := make([]hostNativeArtifact, 0, len(flows)+1)
	for _, flow := range flows {
		name := "runecontext-" + flow.id + ".md"
		artifacts = append(artifacts, hostNativeArtifact{
			relPath: filepath.ToSlash(filepath.Join(".claude", "skills", name)),
			tool:    "claude-code",
			kind:    "flow_asset",
			id:      "runecontext:" + flow.id,
			shim:    false,
			content: buildHostNativeFlowAssetContent("claude-code", flow),
		})
	}
	artifacts = append(artifacts, hostNativeArtifact{
		relPath: filepath.ToSlash(filepath.Join(".claude", "commands", "runecontext.md")),
		tool:    "claude-code",
		kind:    "discoverability_shim",
		id:      "runecontext:index",
		shim:    true,
		content: buildClaudeCommandIndexShimContent(flows),
	})
	sortHostNativeArtifacts(artifacts)
	return artifacts
}

func buildCodexHostNativeArtifacts() []hostNativeArtifact {
	flows := toolFlowMappings("codex")
	artifacts := make([]hostNativeArtifact, 0, len(flows))
	for _, flow := range flows {
		name := "runecontext-" + flow.id + ".md"
		artifacts = append(artifacts, hostNativeArtifact{
			relPath: filepath.ToSlash(filepath.Join(".agents", "skills", name)),
			tool:    "codex",
			kind:    "flow_asset",
			id:      "runecontext:" + flow.id,
			shim:    false,
			content: buildHostNativeFlowAssetContent("codex", flow),
		})
	}
	sortHostNativeArtifacts(artifacts)
	return artifacts
}

func sortHostNativeArtifacts(artifacts []hostNativeArtifact) {
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].relPath < artifacts[j].relPath })
}

func toolFlowMappings(tool string) []hostNativeFlow {
	return []hostNativeFlow{
		{
			id:          "change-new",
			name:        "change new",
			description: "Create a new RuneContext change",
			source:      "adapters/" + tool + "/flows/change-new.md",
		},
		{
			id:          "change-shape",
			name:        "change shape",
			description: "Shape an existing RuneContext change",
			source:      "adapters/" + tool + "/flows/change-shape.md",
		},
		{
			id:          "standard-discover",
			name:        "standard discover",
			description: "Discover standards candidates for promotion",
			source:      "adapters/" + tool + "/flows/standard-discover.md",
		},
		{
			id:          "promote",
			name:        "promote",
			description: "Advance RuneContext promotion state",
			source:      "adapters/" + tool + "/flows/promote.md",
		},
	}
}

func buildHostNativeFlowAssetContent(tool string, flow hostNativeFlow) []byte {
	body := buildHostNativeBody(tool, flow.id, hostNativeKindFlowAsset)
	lines := append(hostNativeFrontmatter(tool, flow, hostNativeKindFlowAsset), []string{
		"<!-- " + hostNativeOwnershipMarker + " -->",
		"<!-- runecontext-tool: " + tool + " -->",
		"<!-- runecontext-kind: flow_asset -->",
		"<!-- runecontext-id: runecontext:" + flow.id + " -->",
		"# RuneContext Skill: " + flow.name,
		"",
		body,
	}...)
	return []byte(strings.Join(lines, "\n") + "\n")
}

func buildHostNativeCommandShimContent(tool string, flow hostNativeFlow) []byte {
	body := buildHostNativeBody(tool, flow.id, hostNativeKindDiscoverabilityShim)
	lines := append(hostNativeFrontmatter(tool, flow, hostNativeKindDiscoverabilityShim), []string{
		"<!-- " + hostNativeOwnershipMarker + " -->",
		"<!-- runecontext-tool: " + tool + " -->",
		"<!-- runecontext-kind: discoverability_shim -->",
		"<!-- runecontext-id: runecontext:" + flow.id + " -->",
		"# RuneContext Command Shim: " + flow.name,
		"",
		body,
	}...)
	return []byte(strings.Join(lines, "\n") + "\n")
}

func buildClaudeCommandIndexShimContent(flows []hostNativeFlow) []byte {
	lines := append(hostNativeIndexFrontmatter("claude-code"), []string{
		"<!-- " + hostNativeOwnershipMarker + " -->",
		"<!-- runecontext-tool: claude-code -->",
		"<!-- runecontext-kind: discoverability_shim -->",
		"<!-- runecontext-id: runecontext:index -->",
		"# RuneContext Command Shim Index",
		"",
		"This file is a discoverability shim. Canonical flow assets live in `.claude/skills/`.",
		"",
		"- Adapter role: discoverability shim",
		"",
		"## Commands",
	}...)
	for _, flow := range flows {
		commandPath := strings.ReplaceAll(flow.id, "-", " ")
		lines = append(lines,
			"",
			"- `runecontext:"+flow.id+"`",
			"  - Canonical flow source: `"+flow.source+"`",
			"  - Skill file: `.claude/skills/runecontext-"+flow.id+".md`",
			"",
			"- command_path: `"+commandPath+"`",
		)
	}
	lines = append(lines,
		"",
		buildHostNativeBody("claude-code", "index", hostNativeKindDiscoverabilityShim),
	)
	return []byte(strings.Join(lines, "\n") + "\n")
}

func hostNativeFrontmatter(tool string, flow hostNativeFlow, kind string) []string {
	name := namespacedHostNativeName(flow.id)
	description := flow.description
	if kind == hostNativeKindDiscoverabilityShim {
		description = flow.description
	}
	switch tool {
	case "opencode":
		return []string{
			"---",
			"description: " + description,
			"---",
		}
	case "claude-code", "codex":
		return []string{
			"---",
			"name: " + name,
			"description: " + description,
			"---",
		}
	default:
		return nil
	}
}

func hostNativeIndexFrontmatter(tool string) []string {
	if tool != "claude-code" {
		return nil
	}
	return []string{
		"---",
		"name: runecontext",
		"description: RuneContext discoverability shim index",
		"---",
	}
}

func namespacedHostNativeName(operation string) string {
	return "runecontext-" + operation
}

func buildHostNativeBody(tool, operation, role string) string {
	role = normalizeHostNativeRole(role)
	if supportsShellInjection(tool) {
		return "!`runectx adapter render-host-native --role " + role + " " + tool + " " + operation + "`"
	}
	req := adapterRenderRequest{tool: tool, operation: operation, role: role}
	body, err := renderHostNativeOperationMarkdown(req)
	if err != nil {
		return "- render_error: `" + strings.ReplaceAll(err.Error(), "`", "") + "`"
	}
	return strings.TrimSpace(body)
}

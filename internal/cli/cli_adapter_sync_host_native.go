package cli

import (
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
	id      string
	name    string
	command string
	source  string
}

func buildHostNativeArtifacts(tool string) ([]hostNativeArtifact, error) {
	switch tool {
	case "opencode":
		return buildOpenCodeHostNativeArtifacts(), nil
	case "claude-code":
		return buildClaudeCodeHostNativeArtifacts(), nil
	case "codex":
		return buildCodexHostNativeArtifacts(), nil
	default:
		return nil, nil
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
			id:      "change-new",
			name:    "change new",
			command: "runectx change new --title \"<title>\" --type <type> [--size <size>] [--shape <minimum|full>] [--bundle <bundle-id>] [--description \"<text>\"] [--path <project-root>]",
			source:  "adapters/" + tool + "/flows/change-new.md",
		},
		{
			id:      "change-shape",
			name:    "change shape",
			command: "runectx change shape CHANGE_ID [--design <text>] [--verification <text>] [--task <text>] [--reference <text>] [--path <project-root>]",
			source:  "adapters/" + tool + "/flows/change-shape.md",
		},
		{
			id:      "standard-discover",
			name:    "standard discover",
			command: "runectx standard discover [--path <project-root>] [--change <CHANGE_ID>] [--scope-path <path>] [--focus \"<text>\"] [--confirm-handoff] [--target <TYPE:PATH>]",
			source:  "adapters/" + tool + "/flows/standard-discover.md",
		},
		{
			id:      "promote",
			name:    "promote",
			command: "runectx promote CHANGE_ID [--accept|--complete] [--target <TYPE:PATH>] [--path <project-root>]",
			source:  "adapters/" + tool + "/flows/promote.md",
		},
	}
}

func buildHostNativeFlowAssetContent(tool string, flow hostNativeFlow) []byte {
	lines := []string{
		"<!-- " + hostNativeOwnershipMarker + " -->",
		"<!-- runecontext-tool: " + tool + " -->",
		"<!-- runecontext-kind: flow_asset -->",
		"<!-- runecontext-id: runecontext:" + flow.id + " -->",
		"# RuneContext Skill: " + flow.name,
		"",
		"This is a RuneContext-managed host-native flow asset.",
		"",
		"- Canonical flow source: `" + flow.source + "`",
		"- Adapter role: canonical flow asset",
		"- Operation identifier: `runecontext:" + flow.id + "`",
		"",
		"```sh",
		flow.command,
		"```",
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func buildHostNativeCommandShimContent(tool string, flow hostNativeFlow) []byte {
	lines := []string{
		"<!-- " + hostNativeOwnershipMarker + " -->",
		"<!-- runecontext-tool: " + tool + " -->",
		"<!-- runecontext-kind: discoverability_shim -->",
		"<!-- runecontext-id: runecontext:" + flow.id + " -->",
		"# RuneContext Command Shim: " + flow.name,
		"",
		"This file is a discoverability shim that points to the canonical flow asset.",
		"",
		"- Canonical flow source: `" + flow.source + "`",
		"- Adapter role: discoverability shim",
		"- Operation identifier: `runecontext:" + flow.id + "`",
		"",
		"```sh",
		flow.command,
		"```",
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func buildClaudeCommandIndexShimContent(flows []hostNativeFlow) []byte {
	lines := []string{
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
	}
	for _, flow := range flows {
		lines = append(lines,
			"",
			"- `runecontext:"+flow.id+"`",
			"  - Canonical flow source: `"+flow.source+"`",
			"  - Skill file: `.claude/skills/runecontext-"+flow.id+".md`",
			"",
			"```sh",
			flow.command,
			"```",
		)
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

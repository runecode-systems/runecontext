package cli

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	hostNativeKindFlowAsset           = "flow_asset"
	hostNativeKindDiscoverabilityShim = "discoverability_shim"
)

type hostNativeOwnershipHeader struct {
	Tool string
	Kind string
	ID   string
}

func requiredOwnershipHeader(artifact hostNativeArtifact) hostNativeOwnershipHeader {
	return hostNativeOwnershipHeader{
		Tool: artifact.tool,
		Kind: artifact.kind,
		ID:   artifact.id,
	}
}

func validateHostNativeOwnershipForWrite(content []byte, rel string, artifact hostNativeArtifact) error {
	parsed, ok := parseHostNativeOwnershipHeader(content)
	if !ok {
		return fmt.Errorf("host-native artifact conflict at %s: existing file is not RuneContext-managed", rel)
	}
	required := requiredOwnershipHeader(artifact)
	if parsed != required {
		return fmt.Errorf("host-native artifact conflict at %s: ownership header mismatch", rel)
	}
	return nil
}

func validateHostNativeOwnershipForDelete(content []byte, rel, tool string) error {
	parsed, ok := parseHostNativeOwnershipHeader(content)
	if !ok {
		return fmt.Errorf("host-native artifact conflict at %s: existing file is not RuneContext-managed", rel)
	}
	if parsed.Tool != tool {
		return fmt.Errorf("host-native artifact conflict at %s: ownership header tool mismatch", rel)
	}
	if !ownershipMatchesDeletePath(rel, parsed) {
		return fmt.Errorf("host-native artifact conflict at %s: ownership header path mismatch", rel)
	}
	return nil
}

func parseHostNativeOwnershipHeader(content []byte) (hostNativeOwnershipHeader, bool) {
	lines := strings.Split(normalizeOwnershipLineEndings(string(content)), "\n")
	start := skipFrontmatter(lines)
	if len(lines[start:]) < 4 {
		return hostNativeOwnershipHeader{}, false
	}
	if lines[start] != "<!-- "+hostNativeOwnershipMarker+" -->" {
		return hostNativeOwnershipHeader{}, false
	}
	tool, ok := parseOwnedCommentLine(lines[start+1], "runecontext-tool: ")
	if !ok || !isSupportedHostNativeTool(tool) {
		return hostNativeOwnershipHeader{}, false
	}
	kind, ok := parseOwnedCommentLine(lines[start+2], "runecontext-kind: ")
	if !ok || !isSupportedHostNativeKind(kind) {
		return hostNativeOwnershipHeader{}, false
	}
	id, ok := parseOwnedCommentLine(lines[start+3], "runecontext-id: ")
	if !ok || !strings.HasPrefix(id, "runecontext:") || len(id) == len("runecontext:") {
		return hostNativeOwnershipHeader{}, false
	}
	return hostNativeOwnershipHeader{Tool: tool, Kind: kind, ID: id}, true
}

func normalizeOwnershipLineEndings(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	return strings.ReplaceAll(content, "\r", "\n")
}

func skipFrontmatter(lines []string) int {
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return 0
	}
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			if idx+1 < len(lines) && strings.TrimSpace(lines[idx+1]) == "" {
				return idx + 2
			}
			return idx + 1
		}
	}
	return 0
}

func parseOwnedCommentLine(line, key string) (string, bool) {
	prefix := "<!-- " + key
	suffix := " -->"
	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, suffix) {
		return "", false
	}
	value := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix))
	if value == "" {
		return "", false
	}
	return value, true
}

func isSupportedHostNativeTool(tool string) bool {
	switch tool {
	case "opencode", "claude-code", "codex":
		return true
	default:
		return false
	}
}

func isSupportedHostNativeKind(kind string) bool {
	return kind == hostNativeKindFlowAsset || kind == hostNativeKindDiscoverabilityShim
}

func ownershipMatchesDeletePath(rel string, header hostNativeOwnershipHeader) bool {
	rel = filepath.ToSlash(rel)
	switch header.Tool {
	case "opencode":
		return ownershipMatchesNamespacedMarkdown(rel, header, ".opencode/skills/", ".opencode/commands/")
	case "claude-code":
		if ownershipMatchesNamespacedMarkdown(rel, header, ".claude/skills/") {
			return true
		}
		if rel == ".claude/commands/runecontext.md" {
			return header.ID == "runecontext:index" && header.Kind == hostNativeKindDiscoverabilityShim
		}
	case "codex":
		return ownershipMatchesNamespacedMarkdown(rel, header, ".agents/skills/")
	}
	return false
}

func ownershipMatchesNamespacedMarkdown(rel string, header hostNativeOwnershipHeader, prefixes ...string) bool {
	if !ownershipPathHasAllowedPrefix(rel, prefixes...) {
		return false
	}
	base := filepath.Base(rel)
	return strings.HasPrefix(base, "runecontext-") && strings.HasSuffix(base, ".md") && strings.HasPrefix(header.ID, "runecontext:")
}

func ownershipPathHasAllowedPrefix(rel string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(rel, prefix) {
			return true
		}
	}
	return false
}

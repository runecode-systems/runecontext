package cli

import (
	"fmt"
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

func validateHostNativeOwnershipForDelete(content []byte, rel string) error {
	if _, ok := parseHostNativeOwnershipHeader(content); !ok {
		return fmt.Errorf("host-native artifact conflict at %s: existing file is not RuneContext-managed", rel)
	}
	return nil
}

func parseHostNativeOwnershipHeader(content []byte) (hostNativeOwnershipHeader, bool) {
	lines := strings.Split(string(content), "\n")
	if len(lines) < 4 {
		return hostNativeOwnershipHeader{}, false
	}
	if lines[0] != "<!-- "+hostNativeOwnershipMarker+" -->" {
		return hostNativeOwnershipHeader{}, false
	}
	tool, ok := parseOwnedCommentLine(lines[1], "runecontext-tool: ")
	if !ok || !isSupportedHostNativeTool(tool) {
		return hostNativeOwnershipHeader{}, false
	}
	kind, ok := parseOwnedCommentLine(lines[2], "runecontext-kind: ")
	if !ok || !isSupportedHostNativeKind(kind) {
		return hostNativeOwnershipHeader{}, false
	}
	id, ok := parseOwnedCommentLine(lines[3], "runecontext-id: ")
	if !ok || !strings.HasPrefix(id, "runecontext:") || len(id) == len("runecontext:") {
		return hostNativeOwnershipHeader{}, false
	}
	return hostNativeOwnershipHeader{Tool: tool, Kind: kind, ID: id}, true
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

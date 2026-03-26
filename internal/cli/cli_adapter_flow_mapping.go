package cli

import "strings"

func flowByOperation(tool, operation string) (hostNativeFlow, bool) {
	operation = sanitizeHostNativeOperation(operation)
	for _, flow := range toolFlowMappings(tool) {
		if flow.id == operation {
			return flow, true
		}
	}
	return hostNativeFlow{}, false
}

func sanitizeHostNativeOperation(operation string) string {
	operation = strings.TrimSpace(strings.ToLower(operation))
	operation = strings.ReplaceAll(operation, "_", "-")
	return operation
}

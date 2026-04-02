package cli

import "strings"

func flowByOperation(tool, operation string) (hostNativeFlow, bool, error) {
	operation = sanitizeHostNativeOperation(operation)
	flows, err := toolFlowMappings(tool)
	if err != nil {
		return hostNativeFlow{}, false, err
	}
	for _, flow := range flows {
		if flow.id == operation {
			return flow, true, nil
		}
	}
	return hostNativeFlow{}, false, nil
}

func sanitizeHostNativeOperation(operation string) string {
	operation = strings.TrimSpace(strings.ToLower(operation))
	operation = strings.ReplaceAll(operation, "_", "-")
	return operation
}

package contracts

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	result := make(map[string]any, len(input))
	for key, value := range input {
		result[key] = cloneTopLevelValue(value)
	}
	return result
}

func nonNilStrings(items []string) []string {
	if items == nil {
		return []string{}
	}
	return items
}

func extractAnySlice(raw any) []any {
	items, _ := raw.([]any)
	if items == nil {
		return []any{}
	}
	return items
}

func intValue(raw any, fallback int) int {
	switch value := raw.(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return fallback
	}
}

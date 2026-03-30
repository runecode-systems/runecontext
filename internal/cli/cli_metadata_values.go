package cli

import "slices"

func noValueSpec() ValueSpec {
	return ValueSpec{Kind: ValueKindNone}
}

func textValueSpec() ValueSpec {
	return ValueSpec{Kind: ValueKindText}
}

func textValueWithSuggestionSpec(provider string) ValueSpec {
	return ValueSpec{Kind: ValueKindText, SuggestionProvider: provider}
}

func enumValueSpec(values ...string) ValueSpec {
	copyValues := append([]string(nil), values...)
	slices.Sort(copyValues)
	return ValueSpec{Kind: ValueKindEnum, EnumValues: copyValues}
}

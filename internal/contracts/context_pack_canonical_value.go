package contracts

func contextPackCanonicalValue(pack *ContextPack) map[string]any {
	value := contextPackDocumentValue(pack)
	delete(value, "pack_hash")
	delete(value, "generated_at")
	return value
}

func contextPackDocumentValue(pack *ContextPack) map[string]any {
	if pack == nil {
		return nil
	}
	return map[string]any{
		"schema_version":       pack.SchemaVersion,
		"canonicalization":     pack.Canonicalization,
		"pack_hash_alg":        pack.PackHashAlg,
		"pack_hash":            pack.PackHash,
		"id":                   pack.ID,
		"requested_bundle_ids": append([]string(nil), pack.RequestedBundleIDs...),
		"resolved_from":        contextPackResolvedFromValue(pack.ResolvedFrom),
		"selected":             contextPackSelectedValue(pack.Selected),
		"excluded":             contextPackExcludedValue(pack.Excluded),
		"generated_at":         pack.GeneratedAt,
	}
}

func contextPackResolvedFromValue(value ContextPackResolvedFrom) map[string]any {
	result := map[string]any{
		"source_mode":         string(value.SourceMode),
		"source_ref":          value.SourceRef,
		"source_verification": string(value.SourceVerification),
		"context_bundle_ids":  append([]string(nil), value.ContextBundleIDs...),
	}
	if value.SourceCommit != "" {
		result["source_commit"] = value.SourceCommit
	}
	if value.VerifiedSignerIdentity != "" {
		result["verified_signer_identity"] = value.VerifiedSignerIdentity
	}
	if value.VerifiedSignerFingerprint != "" {
		result["verified_signer_fingerprint"] = value.VerifiedSignerFingerprint
	}
	return result
}

func contextPackSelectedValue(value ContextPackAspectSet) map[string]any {
	return map[string]any{
		"project":   contextPackSelectedFilesValue(value.Project),
		"standards": contextPackSelectedFilesValue(value.Standards),
		"specs":     contextPackSelectedFilesValue(value.Specs),
		"decisions": contextPackSelectedFilesValue(value.Decisions),
	}
}

func contextPackExcludedValue(value ContextPackExcludedAspectSet) map[string]any {
	return map[string]any{
		"project":   contextPackExcludedFilesValue(value.Project),
		"standards": contextPackExcludedFilesValue(value.Standards),
		"specs":     contextPackExcludedFilesValue(value.Specs),
		"decisions": contextPackExcludedFilesValue(value.Decisions),
	}
}

func contextPackSelectedFilesValue(items []ContextPackSelectedFile) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"path":        item.Path,
			"sha256":      item.SHA256,
			"selected_by": contextPackRuleReferencesValue(item.SelectedBy),
		}
	}
	return result
}

func contextPackExcludedFilesValue(items []ContextPackExcludedFile) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"path":      item.Path,
			"last_rule": contextPackRuleReferenceValue(item.LastRule),
		}
	}
	return result
}

func contextPackRuleReferencesValue(items []ContextPackRuleReference) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = contextPackRuleReferenceValue(item)
	}
	return result
}

func contextPackRuleReferenceValue(item ContextPackRuleReference) map[string]any {
	return map[string]any{
		"bundle":  item.Bundle,
		"aspect":  string(item.Aspect),
		"rule":    string(item.Rule),
		"pattern": item.Pattern,
		"kind":    string(item.Kind),
	}
}

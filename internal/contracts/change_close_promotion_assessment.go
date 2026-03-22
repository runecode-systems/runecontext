package contracts

import (
	"fmt"
	"strings"
)

const (
	promotionSummarySpec     = "Review and promote durable spec updates from this change."
	promotionSummaryStandard = "Review and promote durable standards updates from this change."
	promotionSummaryDecision = "Review and promote durable decision updates from this change."
)

func applyClosePromotionAssessment(updated map[string]any, record *ChangeRecord) {
	if preservePromotionAssessmentState(updated["promotion_assessment"]) {
		return
	}
	targets := closePromotionTargets(record)
	status := "none"
	if len(targets) > 0 {
		status = "suggested"
	}
	updated["promotion_assessment"] = map[string]any{
		"status":            status,
		"suggested_targets": targets,
	}
}

func preservePromotionAssessmentState(raw any) bool {
	promotionRaw, ok := raw.(map[string]any)
	if !ok {
		return false
	}
	status := strings.TrimSpace(fmt.Sprint(promotionRaw["status"]))
	return status == "accepted" || status == "completed"
}

func closePromotionTargets(record *ChangeRecord) []any {
	targets := make([]any, 0)
	for _, path := range sortedUniqueStrings(record.RelatedSpecs) {
		targets = append(targets, map[string]any{
			"target_type": "spec",
			"target_path": path,
			"summary":     promotionSummarySpec,
		})
	}
	if record.Type == "standard" {
		for _, path := range sortedUniqueStrings(record.StandardRefs) {
			targets = append(targets, map[string]any{
				"target_type": "standard",
				"target_path": path,
				"summary":     promotionSummaryStandard,
			})
		}
	}
	for _, path := range sortedUniqueStrings(record.RelatedDecisions) {
		targets = append(targets, map[string]any{
			"target_type": "decision",
			"target_path": path,
			"summary":     promotionSummaryDecision,
		})
	}
	return targets
}

func defaultPromotionTargetSummary(targetType string) string {
	switch strings.TrimSpace(targetType) {
	case "spec":
		return promotionSummarySpec
	case "standard":
		return promotionSummaryStandard
	case "decision":
		return promotionSummaryDecision
	default:
		return ""
	}
}

func trimmedStringField(raw any) string {
	s, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

func sortedUniqueStrings(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	clean := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		clean = append(clean, trimmed)
	}
	if len(clean) == 0 {
		return []string{}
	}
	return uniqueSortedStrings(clean)
}

func closePromotionAssessmentDetails(updated map[string]any) (string, []string) {
	promotionRaw, ok := updated["promotion_assessment"].(map[string]any)
	if !ok {
		return "", nil
	}
	status := strings.TrimSpace(fmt.Sprint(promotionRaw["status"]))
	targets := make([]string, 0)
	for _, targetRaw := range extractAnySlice(promotionRaw["suggested_targets"]) {
		targetMap, ok := targetRaw.(map[string]any)
		if !ok {
			continue
		}
		targetType := trimmedStringField(targetMap["target_type"])
		targetPath := trimmedStringField(targetMap["target_path"])
		if targetType == "" && targetPath == "" {
			continue
		}
		if targetType == "" {
			targets = append(targets, targetPath)
			continue
		}
		targets = append(targets, targetType+":"+targetPath)
	}
	return status, targets
}

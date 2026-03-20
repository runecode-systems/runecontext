package contracts

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func newStatusMap(id, title, changeType, size string, contextBundles []string, now time.Time) map[string]any {
	return map[string]any{
		"schema_version":       1,
		"id":                   id,
		"title":                title,
		"status":               string(StatusProposed),
		"type":                 changeType,
		"size":                 size,
		"verification_status":  "pending",
		"context_bundles":      stringSliceToAny(contextBundles),
		"related_specs":        []any{},
		"related_decisions":    []any{},
		"related_changes":      []any{},
		"depends_on":           []any{},
		"informed_by":          []any{},
		"supersedes":           []any{},
		"superseded_by":        []any{},
		"created_at":           now.Format("2006-01-02"),
		"closed_at":            nil,
		"promotion_assessment": map[string]any{"status": "pending", "suggested_targets": []any{}},
	}
}

func renderStatusYAML(raw map[string]any) ([]byte, error) {
	doc, err := statusDocumentFromMap(raw)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := encodeYAMLDocument(&buf, doc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeYAMLDocument(dst io.Writer, doc any) error {
	encoder := yaml.NewEncoder(dst)
	encoder.SetIndent(2)
	encodeErr := encoder.Encode(doc)
	closeErr := encoder.Close()
	if encodeErr != nil {
		return encodeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}

func statusDocumentFromMap(raw map[string]any) (statusDocument, error) {
	doc := baseStatusDocumentFromMap(raw)
	if err := applyPromotionAssessment(&doc, raw); err != nil {
		return statusDocument{}, err
	}
	applyStatusExtensions(&doc, raw)
	return doc, nil
}

func baseStatusDocumentFromMap(raw map[string]any) statusDocument {
	return statusDocument{
		SchemaVersion:      intValue(raw["schema_version"], 1),
		ID:                 requiredStringValue(raw["id"]),
		Title:              requiredStringValue(raw["title"]),
		Status:             requiredStringValue(raw["status"]),
		Type:               requiredStringValue(raw["type"]),
		Size:               optionalStringValue(raw["size"]),
		VerificationStatus: requiredStringValue(raw["verification_status"]),
		ContextBundles:     nonNilStrings(extractStringList(raw["context_bundles"])),
		RelatedSpecs:       nonNilStrings(extractStringList(raw["related_specs"])),
		RelatedDecisions:   nonNilStrings(extractStringList(raw["related_decisions"])),
		RelatedChanges:     nonNilStrings(extractStringList(raw["related_changes"])),
		DependsOn:          nonNilStrings(extractStringList(raw["depends_on"])),
		InformedBy:         nonNilStrings(extractStringList(raw["informed_by"])),
		Supersedes:         nonNilStrings(extractStringList(raw["supersedes"])),
		SupersededBy:       nonNilStrings(extractStringList(raw["superseded_by"])),
		CreatedAt:          optionalStringValue(raw["created_at"]),
		ClosedAt:           raw["closed_at"],
		PromotionAssessment: promotionAssessmentDocument{
			Status:           "pending",
			SuggestedTargets: []promotionTargetDocument{},
		},
	}
}

func applyPromotionAssessment(doc *statusDocument, raw map[string]any) error {
	promotionRaw, ok := raw["promotion_assessment"].(map[string]any)
	if !ok {
		return nil
	}
	status, err := promotionAssessmentStatusValue(promotionRaw["status"])
	if err != nil {
		return err
	}
	doc.PromotionAssessment.Status = status
	doc.PromotionAssessment.SuggestedTargets = promotionTargetsFromRaw(promotionRaw)
	return nil
}

func promotionTargetsFromRaw(promotionRaw map[string]any) []promotionTargetDocument {
	targets := make([]promotionTargetDocument, 0)
	for _, targetRaw := range extractAnySlice(promotionRaw["suggested_targets"]) {
		targetMap, ok := targetRaw.(map[string]any)
		if !ok {
			continue
		}
		targets = append(targets, promotionTargetDocument{
			TargetType: fmt.Sprint(targetMap["target_type"]),
			TargetPath: fmt.Sprint(targetMap["target_path"]),
			Summary:    fmt.Sprint(targetMap["summary"]),
		})
	}
	return targets
}

func applyStatusExtensions(doc *statusDocument, raw map[string]any) {
	extensions, ok := raw["extensions"].(map[string]any)
	if ok && len(extensions) > 0 {
		doc.Extensions = cloneMap(extensions)
	}
}

func writeStatusMap(path string, raw map[string]any) error {
	data, err := renderStatusYAML(raw)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func promotionAssessmentStatusValue(raw any) (string, error) {
	status := optionalStringValue(raw)
	if status == "" {
		return "pending", nil
	}
	if _, ok := allowedPromotionAssessmentStatuses[status]; !ok {
		return "", fmt.Errorf("promotion_assessment.status must be one of pending, none, suggested, accepted, or completed")
	}
	return status, nil
}

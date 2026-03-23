package contracts

import (
	"fmt"
	"strings"
)

var allowedPromotionTargetTypes = map[string]struct{}{
	"spec":     {},
	"standard": {},
	"decision": {},
}

const allowedPromotionTargetTypeMessage = "spec, standard, decision"

func validatePromoteTargets(targets []string) error {
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			return fmt.Errorf("promotion target cannot be empty")
		}
		targetType, targetPath, ok := strings.Cut(target, ":")
		if !ok || strings.TrimSpace(targetType) == "" || strings.TrimSpace(targetPath) == "" {
			return fmt.Errorf("invalid promotion target %q: expected TYPE:PATH", target)
		}
		typeTrim := strings.TrimSpace(targetType)
		if !allowedPromotionTargetType(typeTrim) {
			return fmt.Errorf("invalid promotion target %q: unknown target type %q (allowed: %s)", target, typeTrim, allowedPromotionTargetTypeMessage)
		}
	}
	return nil
}

func allowedPromotionTargetType(targetType string) bool {
	_, ok := allowedPromotionTargetTypes[targetType]
	return ok
}

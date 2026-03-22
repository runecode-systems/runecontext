package cli

import (
	"fmt"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type standardDiscoverHandoffPlan struct {
	Requested       bool
	Confirmed       bool
	Eligible        bool
	NeedsTarget     bool
	BlockedReason   string
	ChangeID        string
	PromotionTarget string
}

func buildStandardDiscoverHandoffPlan(request standardDiscoverRequest, machine machineOptions, index *contracts.ProjectIndex, candidateTargets []string) standardDiscoverHandoffPlan {
	if request.changeID == "" {
		return standardDiscoverHandoffPlan{}
	}
	plan := standardDiscoverHandoffPlan{
		Requested: request.confirmHandoff || request.changeID != "",
		Confirmed: request.confirmHandoff,
		ChangeID:  request.changeID,
	}
	if !validateStandardDiscoverHandoffEligibility(&plan, request, machine, index, candidateTargets) {
		return plan
	}
	target, ok := resolveRequestedHandoffTarget(request.handoffTarget, candidateTargets)
	if !ok {
		plan.BlockedReason = "target_not_in_candidates"
		return plan
	}
	plan.Eligible = true
	plan.PromotionTarget = target
	return plan
}

func resolveRequestedHandoffTarget(requested string, candidateTargets []string) (string, bool) {
	if len(candidateTargets) == 0 {
		return "", false
	}
	if requested == "" {
		return candidateTargets[0], true
	}
	for _, target := range candidateTargets {
		if target == requested {
			return target, true
		}
	}
	return "", false
}

func validateStandardDiscoverHandoffEligibility(plan *standardDiscoverHandoffPlan, request standardDiscoverRequest, machine machineOptions, index *contracts.ProjectIndex, candidateTargets []string) bool {
	if machine.nonInteractive {
		plan.Confirmed = false
		plan.BlockedReason = "non_interactive_requires_explicit_confirmation"
		return false
	}
	if !request.confirmHandoff {
		plan.BlockedReason = "missing_explicit_confirmation"
		return false
	}
	if index == nil {
		plan.BlockedReason = "missing_project_index"
		return false
	}
	record := index.Changes[request.changeID]
	if record == nil {
		plan.BlockedReason = "change_not_found"
		return false
	}
	if len(candidateTargets) == 0 {
		plan.BlockedReason = "no_candidate_targets"
		return false
	}
	if len(candidateTargets) > 1 && request.handoffTarget == "" {
		plan.NeedsTarget = true
		plan.BlockedReason = "ambiguous_candidate_targets"
		return false
	}
	if !isChangePromotable(record) {
		plan.BlockedReason = "promotion_status_not_suggested"
		return false
	}
	return true
}

func isChangePromotable(record *contracts.ChangeRecord) bool {
	if record == nil {
		return false
	}
	promotionRaw, ok := record.Data["promotion_assessment"].(map[string]any)
	if !ok {
		return false
	}
	status := strings.TrimSpace(fmt.Sprint(promotionRaw["status"]))
	return status == "suggested"
}

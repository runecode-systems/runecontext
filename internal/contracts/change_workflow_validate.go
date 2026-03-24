package contracts

import "fmt"

func validateChangeLifecycleConsistency(index *ProjectIndex) error {
	for _, id := range SortedKeys(index.Changes) {
		if err := validateChangeLifecycleRecord(index.Changes[id]); err != nil {
			return err
		}
	}
	return nil
}

func validateChangeLifecycleRecord(record *ChangeRecord) error {
	if _, ok := lifecycleOrder[record.Status]; !ok {
		return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("unknown lifecycle state %q", record.Status)}
	}
	if err := validateClosedAtConsistency(record); err != nil {
		return err
	}
	if err := validateSupersededByConsistency(record); err != nil {
		return err
	}
	return validateVerificationLifecycleConsistency(record)
}

func validateClosedAtConsistency(record *ChangeRecord) error {
	if isTerminalLifecycleStatus(record.Status) {
		if !record.HasClosedAt {
			return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("terminal change status %q requires closed_at", record.Status)}
		}
		return nil
	}
	if record.HasClosedAt {
		return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("non-terminal change status %q must not set closed_at", record.Status)}
	}
	return nil
}

func validateSupersededByConsistency(record *ChangeRecord) error {
	if record.Status == StatusSuperseded {
		if len(record.SupersededBy) == 0 {
			return &ValidationError{Path: record.StatusPath, Message: "superseded changes must list at least one successor in superseded_by"}
		}
		return nil
	}
	if len(record.SupersededBy) > 0 {
		return &ValidationError{Path: record.StatusPath, Message: "only superseded changes may set superseded_by"}
	}
	return nil
}

func validateVerificationLifecycleConsistency(record *ChangeRecord) error {
	if record.Status == StatusVerified && record.VerificationStatus == "pending" {
		return &ValidationError{Path: record.StatusPath, Message: "verified changes must record a completed verification_status"}
	}
	if isTerminalLifecycleStatus(record.Status) && record.VerificationStatus == "pending" {
		return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("%s changes must not leave verification_status pending", record.Status)}
	}
	return nil
}

func validateRelatedChangeReciprocity(index *ProjectIndex) error {
	for _, id := range SortedKeys(index.Changes) {
		record := index.Changes[id]
		for _, relatedID := range record.RelatedChanges {
			related := index.Changes[relatedID]
			if related != nil && !containsString(related.RelatedChanges, record.ID) {
				return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("related_changes must be reciprocal: %q links to %q but the reverse link is missing", record.ID, relatedID)}
			}
		}
	}
	return nil
}

func validateChangeDependencyCycles(index *ProjectIndex) error {
	if index == nil {
		return nil
	}
	state := map[string]int{}
	for _, id := range SortedKeys(index.Changes) {
		if err := visitChangeDependencyNode(id, index.Changes, state); err != nil {
			return err
		}
	}
	return nil
}

func visitChangeDependencyNode(id string, changes map[string]*ChangeRecord, state map[string]int) error {
	record := changes[id]
	if record == nil {
		return nil
	}
	switch state[id] {
	case 1:
		return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("depends_on relationships contain a cycle involving %q", id)}
	case 2:
		return nil
	}
	state[id] = 1
	for _, dependencyID := range record.DependsOn {
		if _, ok := changes[dependencyID]; !ok {
			continue
		}
		if err := visitChangeDependencyNode(dependencyID, changes, state); err != nil {
			return err
		}
	}
	state[id] = 2
	return nil
}

func validateSupersessionConsistency(index *ProjectIndex) error {
	for _, id := range SortedKeys(index.Changes) {
		if err := validateSuccessorConsistency(index, index.Changes[id]); err != nil {
			return err
		}
		if err := validateSupersedesConsistency(index, index.Changes[id]); err != nil {
			return err
		}
	}
	return nil
}

func validateSuccessorConsistency(index *ProjectIndex, record *ChangeRecord) error {
	for _, successorID := range record.SupersededBy {
		successor := index.Changes[successorID]
		if successor != nil && !containsString(successor.Supersedes, record.ID) {
			return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("superseded_by must be bidirectionally consistent: %q lists %q but the successor does not list %q in supersedes", record.ID, successorID, record.ID)}
		}
	}
	return nil
}

func validateSupersedesConsistency(index *ProjectIndex, record *ChangeRecord) error {
	for _, supersededID := range record.Supersedes {
		superseded := index.Changes[supersededID]
		if superseded == nil {
			continue
		}
		if superseded.Status != StatusSuperseded {
			return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("supersedes references change %q, but that change is not marked superseded", supersededID)}
		}
		if !containsString(superseded.SupersededBy, record.ID) {
			return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("supersedes must be bidirectionally consistent: %q lists %q but the superseded change does not list %q in superseded_by", record.ID, supersededID, record.ID)}
		}
	}
	return nil
}

func validateArtifactTraceabilityConsistency(index *ProjectIndex) error {
	if err := validateChangeArtifactLinks(index); err != nil {
		return err
	}
	if err := validateSpecArtifactLinks(index); err != nil {
		return err
	}
	return validateDecisionArtifactLinks(index)
}

func validateChangeArtifactLinks(index *ProjectIndex) error {
	for _, id := range SortedKeys(index.Changes) {
		record := index.Changes[id]
		if err := validateChangeSpecLinks(index, record); err != nil {
			return err
		}
		if err := validateChangeDecisionLinks(index, record); err != nil {
			return err
		}
	}
	return nil
}

func validateChangeSpecLinks(index *ProjectIndex, record *ChangeRecord) error {
	for _, specPath := range record.RelatedSpecs {
		spec := index.Specs[specPath]
		if spec != nil && !containsString(spec.OriginatingChanges, record.ID) && !containsString(spec.RevisedByChanges, record.ID) {
			return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("related_specs entry %q must point to a spec that references change %q in originating_changes or revised_by_changes", specPath, record.ID)}
		}
	}
	return nil
}

func validateChangeDecisionLinks(index *ProjectIndex, record *ChangeRecord) error {
	for _, decisionPath := range record.RelatedDecisions {
		decision := index.Decisions[decisionPath]
		if decision != nil && !containsString(decision.OriginatingChanges, record.ID) && !containsString(decision.RelatedChanges, record.ID) {
			return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("related_decisions entry %q must point to a decision that references change %q in originating_changes or related_changes", decisionPath, record.ID)}
		}
	}
	return nil
}

func validateSpecArtifactLinks(index *ProjectIndex) error {
	for _, specPath := range SortedKeys(index.Specs) {
		spec := index.Specs[specPath]
		for _, changeID := range append(append([]string{}, spec.OriginatingChanges...), spec.RevisedByChanges...) {
			change := index.Changes[changeID]
			if change != nil && !containsString(change.RelatedSpecs, spec.Path) {
				return &ValidationError{Path: change.StatusPath, Message: fmt.Sprintf("spec %q references change %q, but related_specs on the referenced change's status.yaml is missing %q", spec.Path, changeID, spec.Path)}
			}
		}
	}
	return nil
}

func validateDecisionArtifactLinks(index *ProjectIndex) error {
	for _, decisionPath := range SortedKeys(index.Decisions) {
		decision := index.Decisions[decisionPath]
		for _, changeID := range append(append([]string{}, decision.OriginatingChanges...), decision.RelatedChanges...) {
			change := index.Changes[changeID]
			if change != nil && !containsString(change.RelatedDecisions, decision.Path) {
				return &ValidationError{Path: change.StatusPath, Message: fmt.Sprintf("decision %q references change %q, but related_decisions on the referenced change's status.yaml is missing %q", decision.Path, changeID, decision.Path)}
			}
		}
	}
	return nil
}

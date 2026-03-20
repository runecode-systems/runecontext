package contracts

import "fmt"

func BuildSplitChangeGraph(plan SplitChangePlan) (map[string]ChangeGraphLinks, error) {
	if err := validateSplitChangePlan(plan); err != nil {
		return nil, err
	}
	links, subIDs, err := buildSplitChangeLinks(plan)
	if err != nil {
		return nil, err
	}
	links[plan.UmbrellaID] = ChangeGraphLinks{RelatedChanges: uniqueSortedStrings(append([]string(nil), subIDs...))}
	if err := validateSplitChangeCycles(plan.UmbrellaID, plan.SubChanges); err != nil {
		return nil, err
	}
	return links, nil
}

func validateSplitChangePlan(plan SplitChangePlan) error {
	if plan.UmbrellaID == "" {
		return fmt.Errorf("umbrella change ID is required")
	}
	return validateChangeIDValue(plan.UmbrellaID, "umbrella change ID")
}

func buildSplitChangeLinks(plan SplitChangePlan) (map[string]ChangeGraphLinks, []string, error) {
	links := map[string]ChangeGraphLinks{plan.UmbrellaID: {}}
	seen := map[string]struct{}{plan.UmbrellaID: {}}
	subIDs, err := collectSplitSubIDs(plan, seen)
	if err != nil {
		return nil, nil, err
	}
	for _, sub := range plan.SubChanges {
		dependsOn, err := normalizedSubChangeDependencies(sub)
		if err != nil {
			return nil, nil, err
		}
		links[sub.ID] = ChangeGraphLinks{RelatedChanges: relatedSplitChangeIDs(plan.UmbrellaID, sub.ID, subIDs), DependsOn: dependsOn}
	}
	return links, subIDs, nil
}

func collectSplitSubIDs(plan SplitChangePlan, seen map[string]struct{}) ([]string, error) {
	subIDs := make([]string, 0, len(plan.SubChanges))
	for _, sub := range plan.SubChanges {
		if sub.ID == "" {
			return nil, fmt.Errorf("sub-change ID is required")
		}
		if err := validateChangeIDValue(sub.ID, fmt.Sprintf("sub-change ID %q", sub.ID)); err != nil {
			return nil, err
		}
		if sub.ID == plan.UmbrellaID {
			return nil, fmt.Errorf("sub-change ID %q must differ from umbrella change ID", sub.ID)
		}
		if _, ok := seen[sub.ID]; ok {
			return nil, fmt.Errorf("duplicate split-change ID %q", sub.ID)
		}
		seen[sub.ID] = struct{}{}
		subIDs = append(subIDs, sub.ID)
	}
	return subIDs, nil
}

func normalizedSubChangeDependencies(sub SplitSubChange) ([]string, error) {
	dependsOn := append([]string(nil), sub.DependsOn...)
	for _, depID := range dependsOn {
		if depID == sub.ID {
			return nil, fmt.Errorf("sub-change %q must not depend on itself", sub.ID)
		}
		if err := validateChangeIDValue(depID, fmt.Sprintf("depends_on entry %q", depID)); err != nil {
			return nil, err
		}
	}
	return uniqueSortedStrings(dependsOn), nil
}

func relatedSplitChangeIDs(umbrellaID, currentID string, subIDs []string) []string {
	related := make([]string, 0, len(subIDs))
	related = append(related, umbrellaID)
	for _, otherID := range subIDs {
		if otherID != currentID {
			related = append(related, otherID)
		}
	}
	return uniqueSortedStrings(related)
}

func validateSplitChangeCycles(umbrellaID string, subChanges []SplitSubChange) error {
	adjacency := buildSplitChangeAdjacency(umbrellaID, subChanges)
	state := map[string]int{}
	for _, sub := range subChanges {
		if err := visitSplitChangeNode(sub.ID, adjacency, state); err != nil {
			return err
		}
	}
	return nil
}

func buildSplitChangeAdjacency(umbrellaID string, subChanges []SplitSubChange) map[string][]string {
	adjacency := map[string][]string{}
	for _, sub := range subChanges {
		for _, depID := range uniqueSortedStrings(sub.DependsOn) {
			if depID != umbrellaID {
				adjacency[sub.ID] = append(adjacency[sub.ID], depID)
			}
		}
	}
	return adjacency
}

func visitSplitChangeNode(id string, adjacency map[string][]string, state map[string]int) error {
	switch state[id] {
	case 1:
		return fmt.Errorf("split-change dependencies contain a cycle involving %q", id)
	case 2:
		return nil
	}
	state[id] = 1
	for _, depID := range adjacency[id] {
		if _, ok := adjacency[depID]; ok {
			if err := visitSplitChangeNode(depID, adjacency, state); err != nil {
				return err
			}
		}
	}
	state[id] = 2
	return nil
}

package contracts

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (c *BundleCatalog) linearize(id string) ([]*bundleDefinition, error) {
	state := newBundleLinearizationState()
	if err := c.visitBundleLinearization(id, 1, state); err != nil {
		return nil, err
	}
	return state.ordered, nil
}

func (c *BundleCatalog) linearizeRequest(bundleIDs []string) ([]*bundleDefinition, error) {
	state := newBundleLinearizationState()
	for _, bundleID := range bundleIDs {
		if err := c.visitBundleLinearization(bundleID, 1, state); err != nil {
			return nil, err
		}
	}
	return state.ordered, nil
}

type bundleLinearizationState struct {
	ordered    []*bundleDefinition
	emitted    map[string]struct{}
	stack      []string
	stackIndex map[string]int
}

func newBundleLinearizationState() *bundleLinearizationState {
	return &bundleLinearizationState{ordered: make([]*bundleDefinition, 0), emitted: map[string]struct{}{}, stack: make([]string, 0), stackIndex: map[string]int{}}
}

func (c *BundleCatalog) visitBundleLinearization(bundleID string, depth int, state *bundleLinearizationState) error {
	bundle, err := c.requireKnownBundle(bundleID)
	if err != nil {
		return err
	}
	if bundleAlreadyEmitted(bundleID, state) {
		return nil
	}
	if err := validateBundleLinearizationDepth(bundle, depth); err != nil {
		return err
	}
	if err := validateBundleLinearizationCycle(bundle, bundleID, state); err != nil {
		return err
	}
	pushBundleLinearization(bundleID, state)
	defer popBundleLinearization(bundleID, state)
	for _, parentID := range bundle.Extends {
		if err := c.visitBundleLinearization(parentID, depth+1, state); err != nil {
			return err
		}
	}
	markBundleLinearized(bundle, state)
	return nil
}

func (c *BundleCatalog) requireKnownBundle(bundleID string) (*bundleDefinition, error) {
	bundle, ok := c.bundles[bundleID]
	if !ok {
		return nil, &ValidationError{Path: filepath.Join(c.Root, "bundles"), Message: fmt.Sprintf("unknown bundle %q", bundleID)}
	}
	return bundle, nil
}

func bundleAlreadyEmitted(bundleID string, state *bundleLinearizationState) bool {
	_, ok := state.emitted[bundleID]
	return ok
}

func validateBundleLinearizationDepth(bundle *bundleDefinition, depth int) error {
	if depth > maxBundleInheritanceDepth {
		return &ValidationError{Path: bundle.Path, Message: fmt.Sprintf("bundle inheritance depth exceeds maximum of %d", maxBundleInheritanceDepth)}
	}
	return nil
}

func validateBundleLinearizationCycle(bundle *bundleDefinition, bundleID string, state *bundleLinearizationState) error {
	if idx, ok := state.stackIndex[bundleID]; ok {
		cycle := append(append([]string{}, state.stack[idx:]...), bundleID)
		return &ValidationError{Path: bundle.Path, Message: fmt.Sprintf("bundle inheritance cycle detected: %s", strings.Join(cycle, " -> "))}
	}
	return nil
}

func pushBundleLinearization(bundleID string, state *bundleLinearizationState) {
	state.stackIndex[bundleID] = len(state.stack)
	state.stack = append(state.stack, bundleID)
}

func popBundleLinearization(bundleID string, state *bundleLinearizationState) {
	delete(state.stackIndex, bundleID)
	state.stack = state.stack[:len(state.stack)-1]
}

func markBundleLinearized(bundle *bundleDefinition, state *bundleLinearizationState) {
	state.emitted[bundle.ID] = struct{}{}
	state.ordered = append(state.ordered, bundle)
}

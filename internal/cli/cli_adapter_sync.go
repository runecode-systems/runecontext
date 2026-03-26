package cli

import (
	"fmt"
	"sort"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type adapterSyncState struct {
	absRoot         string
	tool            string
	hostNativeFiles []hostNativeArtifact
	plan            []contracts.FileMutation
}

func buildAdapterSyncState(request adapterRequest) (adapterSyncState, error) {
	absRoot, err := resolveAbsoluteRoot(request.root)
	if err != nil {
		return adapterSyncState{}, err
	}
	hostNativeFiles, err := buildHostNativeArtifacts(request.tool)
	if err != nil {
		return adapterSyncState{}, err
	}
	if len(hostNativeFiles) == 0 {
		return adapterSyncState{}, fmt.Errorf("adapter %q does not define repo-local host-native artifacts", request.tool)
	}
	plan, err := buildAdapterSyncPlan(absRoot, request.tool, hostNativeFiles)
	if err != nil {
		return adapterSyncState{}, err
	}
	return adapterSyncState{
		absRoot:         absRoot,
		tool:            request.tool,
		hostNativeFiles: hostNativeFiles,
		plan:            plan,
	}, nil
}

func buildAdapterSyncPlan(absRoot, tool string, hostNativeFiles []hostNativeArtifact) ([]contracts.FileMutation, error) {
	plan, err := plannedHostNativeWrites(absRoot, hostNativeFiles)
	if err != nil {
		return nil, err
	}
	hostDeletes, err := plannedHostNativeDeletesFromExisting(absRoot, tool, hostNativeFiles)
	if err != nil {
		return nil, err
	}
	plan = append(plan, hostDeletes...)
	sort.Slice(plan, func(i, j int) bool {
		if plan[i].Path == plan[j].Path {
			return plan[i].Action < plan[j].Action
		}
		return plan[i].Path < plan[j].Path
	})
	return plan, nil
}

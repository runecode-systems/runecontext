package cli

import (
	"fmt"
	"io"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

const standardDiscoverCommand = "standard_discover"

type standardDiscoverRequest struct {
	root           string
	explicitRoot   bool
	changeID       string
	confirmHandoff bool
	handoffTarget  string
}

func runStandard(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("standard", standardUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if len(remaining) == 0 {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("standard", standardUsage, fmt.Errorf("standard subcommand is required")), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if isHelpToken(remaining[0]) {
		if len(remaining) != 1 {
			emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("standard", standardUsage, fmt.Errorf("help does not accept additional arguments")), machine), exitUsage, failureClassUsage)
			return exitUsage
		}
		emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", "standard"}, {"usage", standardUsage}}, machine), exitOK, failureClassNone)
		return exitOK
	}
	switch remaining[0] {
	case "discover":
		return runStandardDiscover(remaining[1:], machine, stdout, stderr)
	default:
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("standard", standardUsage, fmt.Errorf("unknown standard subcommand %q", remaining[0])), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
}

func runStandardDiscover(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	request, err := parseStandardDiscoverArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines(standardDiscoverCommand, standardDiscoverUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, standardDiscoverCommand, machine)
	if code != exitOK {
		return code
	}
	defer project.close()

	index, err := project.validator.ValidateLoadedProject(project.loaded)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(standardDiscoverCommand, project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	defer index.Close()

	candidateStandards := discoverCandidateStandards(index)
	candidateTargets := standardPromotionTargets(candidateStandards)
	handoff := buildStandardDiscoverHandoffPlan(request, machine, index, candidateTargets)
	output := buildStandardDiscoverOutput(project.absRoot, project.loaded, candidateStandards, candidateTargets, handoff)
	if machine.explain {
		output = appendStandardDiscoverExplainLines(output, project.loaded, candidateStandards, candidateTargets, handoff)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func parseStandardDiscoverArgs(args []string) (standardDiscoverRequest, error) {
	request := standardDiscoverRequest{root: "."}
	positionals := make([]string, 0, len(args))
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		case "--change":
			return assignStringFlag(args, flag, &request.changeID)
		case "--confirm-handoff":
			if flag.hasValue {
				return flag.next, fmt.Errorf("--confirm-handoff does not accept a value")
			}
			request.confirmHandoff = true
			return flag.next, nil
		case "--target":
			return assignStringFlag(args, flag, &request.handoffTarget)
		default:
			return flag.next, fmt.Errorf("unknown standard discover flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return standardDiscoverRequest{}, err
	}
	if len(positionals) > 0 {
		return standardDiscoverRequest{}, fmt.Errorf("standard discover does not accept positional arguments")
	}
	if (request.confirmHandoff || request.handoffTarget != "") && request.changeID == "" {
		return standardDiscoverRequest{}, fmt.Errorf("--confirm-handoff and --target require --change")
	}
	return request, nil
}

func discoverCandidateStandards(index *contracts.ProjectIndex) []string {
	if index == nil {
		return nil
	}
	paths := contracts.SortedKeys(index.Standards)
	candidates := make([]string, 0, len(paths))
	for _, path := range paths {
		record := index.Standards[path]
		if record == nil || record.Status != contracts.StandardStatusActive {
			continue
		}
		candidates = append(candidates, path)
	}
	return candidates
}

func standardPromotionTargets(paths []string) []string {
	targets := make([]string, 0, len(paths))
	for _, path := range paths {
		targets = append(targets, "standard:"+path)
	}
	return targets
}

func buildStandardDiscoverOutput(absRoot string, loaded *contracts.LoadedProject, candidateStandards, candidateTargets []string, handoff standardDiscoverHandoffPlan) []line {
	output := []line{
		{"result", "ok"},
		{"command", standardDiscoverCommand},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"mutation_performed", "false"},
		{"candidate_standard_count", fmt.Sprintf("%d", len(candidateStandards))},
	}
	if loaded != nil && loaded.Resolution != nil {
		output = append(output,
			line{"project_root", loaded.Resolution.ProjectRoot},
			line{"source_root", loaded.Resolution.SourceRoot},
			line{"source_mode", string(loaded.Resolution.SourceMode)},
		)
	}
	output = appendStringItems(output, "candidate_standard", candidateStandards)
	output = append(output, line{"candidate_promotion_target_count", fmt.Sprintf("%d", len(candidateTargets))})
	output = appendStringItems(output, "candidate_promotion_target", candidateTargets)
	output = append(output, line{"handoff_requested", boolString(handoff.Requested)})
	output = append(output, line{"handoff_confirmed", boolString(handoff.Confirmed)})
	output = append(output, line{"handoff_eligible", boolString(handoff.Eligible)})
	output = append(output, line{"handoff_target_required", boolString(handoff.NeedsTarget)})
	if handoff.ChangeID != "" {
		output = append(output, line{"handoff_change_id", handoff.ChangeID})
	}
	if handoff.PromotionTarget != "" {
		output = append(output, line{"handoff_promotion_target", handoff.PromotionTarget})
	}
	if handoff.BlockedReason != "" {
		output = append(output, line{"handoff_blocked_reason", handoff.BlockedReason})
	}
	return output
}

func appendStandardDiscoverExplainLines(lines []line, loaded *contracts.LoadedProject, candidateStandards, candidateTargets []string, handoff standardDiscoverHandoffPlan) []line {
	lines = append(lines,
		line{"explain_scope", "standards-discovery,promotion-handoff"},
		line{"explain_advisory_only", "true"},
		line{"explain_candidate_standard_count_reason", fmt.Sprintf("%d active standards discovered from the validated project index", len(candidateStandards))},
		line{"explain_candidate_promotion_target_count_reason", fmt.Sprintf("%d promotion targets emitted for downstream runectx promote usage", len(candidateTargets))},
	)
	if handoff.Requested {
		lines = append(lines,
			line{"explain_handoff_requested", "true"},
			line{"explain_handoff_confirmed", boolString(handoff.Confirmed)},
			line{"explain_handoff_eligible", boolString(handoff.Eligible)},
			line{"explain_handoff_target_required", boolString(handoff.NeedsTarget)},
		)
		if handoff.BlockedReason != "" {
			lines = append(lines, line{"explain_handoff_blocked_reason", handoff.BlockedReason})
		}
	}
	if loaded != nil && loaded.Resolution != nil {
		lines = append(lines,
			line{"explain_resolution_selected_config_path", loaded.Resolution.SelectedConfigPath},
			line{"explain_resolution_source_mode", string(loaded.Resolution.SourceMode)},
		)
	}
	return lines
}

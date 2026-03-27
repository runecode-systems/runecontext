package cli

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

const (
	bundleResolveCommand = "bundle_resolve"
	doctorCommand        = "doctor"
)

type bundleResolveRequest struct {
	root         string
	explicitRoot bool
	bundleIDs    []string
}

type doctorRequest struct {
	root         string
	explicitRoot bool
}

func runBundle(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("bundle", bundleUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if len(remaining) == 0 {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("bundle", bundleUsage, fmt.Errorf("bundle subcommand is required")), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if isHelpToken(remaining[0]) {
		if len(remaining) != 1 {
			emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("bundle", bundleUsage, fmt.Errorf("help does not accept additional arguments")), machine), exitUsage, failureClassUsage)
			return exitUsage
		}
		emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", "bundle"}, {"usage", bundleUsage}}, machine), exitOK, failureClassNone)
		return exitOK
	}
	switch remaining[0] {
	case "resolve":
		return runBundleResolve(remaining[1:], machine, stdout, stderr)
	default:
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("bundle", bundleUsage, fmt.Errorf("unknown bundle subcommand %q", remaining[0])), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
}

func runBundleResolve(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	request, err := parseBundleResolveArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines(bundleResolveCommand, bundleResolveUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, bundleResolveCommand, machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	index, err := project.validator.ValidateLoadedProject(project.loaded)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(bundleResolveCommand, project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	defer index.Close()
	resolution, err := index.Bundles.ResolveRequest(request.bundleIDs)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(bundleResolveCommand, project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	report, err := buildBundleResolveContextPackReport(index, request.bundleIDs, machine.explain)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(bundleResolveCommand, project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	diagnostics := collectDiagnostics(index)
	output := buildBundleResolveOutput(project.absRoot, project.loaded, request, resolution, report, diagnostics)
	if machine.explain {
		output = appendBundleResolveExplainLines(output, project.loaded, resolution, report, diagnostics)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func buildBundleResolveContextPackReport(index *contracts.ProjectIndex, bundleIDs []string, explain bool) (*contracts.ContextPackReport, error) {
	return index.BuildContextPackReport(contracts.ContextPackReportOptions{
		ContextPackOptions: contracts.ContextPackOptions{
			BundleIDs:   append([]string(nil), bundleIDs...),
			GeneratedAt: time.Now().UTC().Truncate(time.Second),
		},
		Explain: explain,
	})
}

func parseBundleResolveArgs(args []string) (bundleResolveRequest, error) {
	request := bundleResolveRequest{root: "."}
	positionals := make([]string, 0, len(args))
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		if flag.name != "--path" {
			return flag.next, fmt.Errorf("unknown bundle resolve flag %q", flag.raw)
		}
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return bundleResolveRequest{}, err
	}
	return finalizeBundleResolveRequest(request, positionals)
}

func finalizeBundleResolveRequest(request bundleResolveRequest, positionals []string) (bundleResolveRequest, error) {
	if len(positionals) == 0 {
		return bundleResolveRequest{}, fmt.Errorf("bundle resolve requires at least one bundle ID")
	}
	ids := make([]string, 0, len(positionals))
	for _, positional := range positionals {
		trimmed := strings.TrimSpace(positional)
		if trimmed == "" {
			continue
		}
		ids = append(ids, trimmed)
	}
	if len(ids) == 0 {
		return bundleResolveRequest{}, fmt.Errorf("bundle resolve requires at least one bundle ID")
	}
	request.bundleIDs = ids
	return request, nil
}

func runDoctor(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines(doctorCommand, doctorUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	request, err := parseDoctorArgs(remaining)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines(doctorCommand, doctorUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, doctorCommand, machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	index, err := project.validator.ValidateLoadedProject(project.loaded)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(doctorCommand, project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	defer index.Close()
	upgradePlan, err := buildUpgradeReadinessFromIndex(project.absRoot, index)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(doctorCommand, project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	diagnostics := collectDiagnostics(index)
	diagnostics = append(diagnostics, upgradePlanDiagnostics(upgradePlan)...)
	warnings := doctorEnvironmentWarnings()
	output := buildDoctorOutput(project.absRoot, project.loaded, diagnostics, warnings, upgradePlan)
	if machine.explain {
		output = appendDoctorExplainLines(output, project.loaded, diagnostics, warnings, upgradePlan)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func parseDoctorArgs(args []string) (doctorRequest, error) {
	request := doctorRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		if flag.name != "--path" {
			return flag.next, fmt.Errorf("unknown doctor flag %q", flag.raw)
		}
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return doctorRequest{}, err
	}
	if len(positionals) > 1 {
		return doctorRequest{}, fmt.Errorf("expected at most one path argument")
	}
	if len(positionals) == 1 {
		if request.explicitRoot {
			return doctorRequest{}, fmt.Errorf("cannot use both --path and a positional path argument")
		}
		request.root = positionals[0]
		request.explicitRoot = true
	}
	return request, nil
}

func buildDoctorOutput(absRoot string, loaded *contracts.LoadedProject, diagnostics []emittedDiagnostic, warnings []string, upgradePlan upgradePlan) []line {
	output := []line{
		{"result", "ok"},
		{"command", doctorCommand},
		{"root", absRoot},
		{"cli_version", normalizedRunecontextVersion()},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"project_runecontext_version", upgradePlan.CurrentVersion},
		{"upgrade_target_version", upgradePlan.TargetVersion},
		{"upgrade_state", string(upgradePlan.State)},
		{"upgrade_network_access", boolString(upgradePlan.NetworkAccess)},
	}
	if loaded != nil && loaded.Resolution != nil {
		output = append(output,
			line{"project_root", loaded.Resolution.ProjectRoot},
			line{"source_root", loaded.Resolution.SourceRoot},
			line{"source_mode", string(loaded.Resolution.SourceMode)},
			line{"source_ref", loaded.Resolution.SourceRef},
			line{"verification_posture", string(loaded.Resolution.VerificationPosture)},
		)
	}
	output = append(output, line{"diagnostic_count", fmt.Sprintf("%d", len(diagnostics))})
	output = appendValidateDiagnosticLines(output, diagnostics)
	output = appendStringItems(output, "upgrade_next_action", upgradePlan.NextActions)
	output = appendStringItems(output, "upgrade_conflict", upgradePlan.Conflicts)
	return appendWarnings(output, warnings)
}

func appendDoctorExplainLines(lines []line, loaded *contracts.LoadedProject, diagnostics []emittedDiagnostic, warnings []string, upgradePlan upgradePlan) []line {
	lines = append(lines,
		line{"explain_scope", "resolution,environment,upgrade-readiness"},
		line{"explain_diagnostic_count", fmt.Sprintf("%d", len(diagnostics))},
		line{"explain_warning_count", fmt.Sprintf("%d", len(warnings))},
		line{"explain_upgrade_state", string(upgradePlan.State)},
	)
	if loaded != nil && loaded.Resolution != nil {
		lines = append(lines,
			line{"explain_resolution_selected_config_path", loaded.Resolution.SelectedConfigPath},
			line{"explain_resolution_verification_posture", string(loaded.Resolution.VerificationPosture)},
		)
	}
	return lines
}

func doctorEnvironmentWarnings() []string {
	warnings := make([]string, 0, 1)
	if _, err := exec.LookPath("git"); err != nil {
		warnings = append(warnings, "git executable not found in PATH; bundle resolution and change commands may be unavailable")
	}
	return warnings
}

package cli

import (
	"fmt"
	"io"
	"time"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type assuranceCaptureRequest struct {
	root         string
	explicitRoot bool
	bundleIDs    []string
}

func runAssuranceCaptureWithMachine(args []string, stdout, stderr io.Writer, machine machineOptions) int {
	request, err := parseAssuranceCaptureArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance capture", assuranceCaptureUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "assurance capture", machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	if err := ensureAssuranceCaptureTier(project); err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("assurance capture", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	if machine.dryRun {
		emitAssuranceCaptureDryRun(stdout, stderr, machine, project, request)
		return exitOK
	}
	result, err := executeAssuranceCaptureContextPack(project, machine, request)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("assurance capture", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildAssuranceCaptureSuccessOutput(project, request.bundleIDs, result)
	output = appendCaptureExplainLines(output, machine.explain)
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	if !machine.jsonOutput {
		fmt.Fprintf(stderr, "Captured context-pack receipt at %s\n", result.ReceiptPath)
	}
	return exitOK
}

func ensureAssuranceCaptureTier(project *cliProject) error {
	if contracts.IsVerifiedAssuranceTierForCLI(project.loaded) {
		return nil
	}
	return fmt.Errorf("assurance_tier must be verified before running assurance capture")
}

func emitAssuranceCaptureDryRun(stdout, stderr io.Writer, machine machineOptions, project *cliProject, request assuranceCaptureRequest) {
	output := []line{
		{"result", "ok"},
		{"command", "assurance capture"},
		{"root", project.absRoot},
		{"selected_config_path", selectedConfigPath(project.loaded)},
		{"mode", "context-pack"},
		{"plan_action_1", "write assurance receipt under assurance/receipts/context-packs/"},
		{"requested_bundle_count", fmt.Sprintf("%d", len(request.bundleIDs))},
	}
	output = appendStringItems(output, "requested_bundle", request.bundleIDs)
	output = appendCaptureExplainLines(output, machine.explain)
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	if !machine.jsonOutput {
		fmt.Fprintln(stderr, "Dry run: would capture context-pack assurance receipt")
	}
}

func buildAssuranceCaptureSuccessOutput(project *cliProject, bundleIDs []string, result *contracts.AssuranceCaptureResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", "assurance capture"},
		{"root", project.absRoot},
		{"selected_config_path", selectedConfigPath(project.loaded)},
		{"receipt_path", result.ReceiptPath},
		{"receipt_id", result.ReceiptID},
		{"pack_hash", result.PackHash},
		{"requested_bundle_count", fmt.Sprintf("%d", len(bundleIDs))},
	}
	output = appendStringItems(output, "requested_bundle", bundleIDs)
	return appendChangedFiles(output, result.ChangedFiles)
}

func appendCaptureExplainLines(output []line, include bool) []line {
	if !include {
		return output
	}
	return append(output,
		line{"explain_scope", "assurance-capture-context-pack"},
		line{"explain_receipt_family", "context-packs"},
		line{"explain_capture_snapshot", "pack and receipt emitted from one validated snapshot"},
	)
}

func parseAssuranceCaptureArgs(args []string) (assuranceCaptureRequest, error) {
	if len(args) == 0 {
		return assuranceCaptureRequest{}, fmt.Errorf("assurance capture requires a subject; expected \"context-pack\"")
	}
	if args[0] != "context-pack" {
		return assuranceCaptureRequest{}, fmt.Errorf("assurance capture supports only \"context-pack\", got %q", args[0])
	}
	request := assuranceCaptureRequest{root: "."}
	positionals := make([]string, 0, len(args)-1)
	err := consumeArgs(args[1:], func(flag parsedFlag) (int, error) {
		if flag.name != "--path" {
			return flag.next, fmt.Errorf("unknown assurance capture flag %q", flag.raw)
		}
		return assignRootFlag(args[1:], flag, &request.root, &request.explicitRoot)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return assuranceCaptureRequest{}, err
	}
	if len(positionals) == 0 {
		return assuranceCaptureRequest{}, fmt.Errorf("assurance capture context-pack requires at least one bundle ID")
	}
	request.bundleIDs = make([]string, 0, len(positionals))
	for _, id := range positionals {
		if id == "" {
			continue
		}
		request.bundleIDs = append(request.bundleIDs, id)
	}
	if len(request.bundleIDs) == 0 {
		return assuranceCaptureRequest{}, fmt.Errorf("assurance capture context-pack requires at least one bundle ID")
	}
	return request, nil
}

func executeAssuranceCaptureContextPack(project *cliProject, _ machineOptions, request assuranceCaptureRequest) (*contracts.AssuranceCaptureResult, error) {
	if project == nil || project.validator == nil || project.loaded == nil {
		return nil, fmt.Errorf("assurance capture requires a loaded project")
	}
	index, err := project.validator.ValidateLoadedProject(project.loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()
	return contracts.CaptureContextPackAssurance(project.validator, project.loaded, index, contracts.ContextPackAssuranceCaptureOptions{
		BundleIDs:   append([]string(nil), request.bundleIDs...),
		GeneratedAt: time.Now().UTC().Truncate(time.Second),
	})
}

package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

const (
	exitOK      = 0
	exitInvalid = 1
	exitUsage   = 2
)

const (
	validateUsage         = "runectx validate [--ssh-allowed-signers PATH] [path]"
	statusUsage           = "runectx status [path]"
	changeUsage           = "runectx change <new|shape|close|reallocate> ..."
	changeNewUsage        = "runectx change new --title TITLE --type TYPE [--size SIZE] [--bundle ID] [--shape minimum|full] [--description TEXT] [--path PATH]"
	changeShapeUsage      = "runectx change shape CHANGE_ID [--design TEXT] [--verification TEXT] [--task TEXT] [--reference TEXT] [--path PATH]"
	changeCloseUsage      = "runectx change close CHANGE_ID [--verification-status STATUS] [--superseded-by ID] [--closed-at YYYY-MM-DD] [--path PATH]"
	changeReallocateUsage = "runectx change reallocate CHANGE_ID [--path PATH]"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return exitOK
	}

	switch args[0] {
	case "validate":
		return runValidate(args[1:], stdout, stderr)
	case "status":
		return runStatus(args[1:], stdout, stderr)
	case "change":
		return runChange(args[1:], stdout, stderr)
	case "help", "--help", "-h":
		printUsage(stdout)
		return exitOK
	default:
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"command", args[0]},
			line{"error_message", fmt.Sprintf("unknown command %q", args[0])},
			line{"usage", validateUsage},
		)
		return exitUsage
	}
}

func runValidate(args []string, stdout, stderr io.Writer) int {
	request, err := parseValidateArgs(args)
	if err != nil {
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"command", "validate"},
			line{"error_message", err.Error()},
			line{"usage", validateUsage},
		)
		return exitUsage
	}

	root := request.root
	resolveOptions := contracts.ResolveOptions{
		ConfigDiscovery: contracts.ConfigDiscoveryNearestAncestor,
		ExecutionMode:   contracts.ExecutionModeLocal,
		GitTrust:        request.gitTrust,
	}
	if request.explicitRoot {
		resolveOptions.ConfigDiscovery = contracts.ConfigDiscoveryExplicitRoot
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"command", "validate"},
			line{"error_message", fmt.Sprintf("failed to resolve path %q: %v", root, err)},
			line{"usage", validateUsage},
		)
		return exitUsage
	}

	schemaRoot, err := locateSchemaRoot()
	if err != nil {
		writeLines(stderr,
			line{"result", "invalid"},
			line{"command", "validate"},
			line{"root", absRoot},
			line{"error_message", err.Error()},
		)
		return exitInvalid
	}

	validator := contracts.NewValidator(schemaRoot)
	index, err := validator.ValidateProjectWithOptions(absRoot, resolveOptions)
	if err != nil {
		lines := []line{
			{"result", "invalid"},
			{"command", "validate"},
			{"root", absRoot},
		}
		var signedTagErr *contracts.SignedTagVerificationError
		if errors.As(err, &signedTagErr) {
			if signedTagErr.Path != "" {
				lines = append(lines, line{"error_path", signedTagErr.Path})
			}
			if signedTagErr.Tag != "" {
				lines = append(lines, line{"error_tag", signedTagErr.Tag})
			}
			lines = append(lines,
				line{"error_reason", string(signedTagErr.Reason)},
				line{"error_message", signedTagErr.Message},
				line{"diagnostic_count", fmt.Sprintf("%d", len(signedTagErr.Diagnostics))},
			)
			if signedTagErr.ResolvedCommit != "" {
				lines = append(lines, line{"resolved_commit", signedTagErr.ResolvedCommit})
			}
			if signedTagErr.SignerIdentity != "" {
				lines = append(lines, line{"verified_signer_identity", signedTagErr.SignerIdentity})
			}
			if signedTagErr.SignerFingerprint != "" {
				lines = append(lines, line{"verified_signer_fingerprint", signedTagErr.SignerFingerprint})
			}
			for i, diagnostic := range signedTagErr.Diagnostics {
				prefix := fmt.Sprintf("diagnostic_%d", i+1)
				lines = append(lines,
					line{prefix + "_severity", string(diagnostic.Severity)},
					line{prefix + "_code", diagnostic.Code},
					line{prefix + "_message", diagnostic.Message},
				)
			}
			writeLines(stderr, lines...)
			return exitInvalid
		}
		var validationErr *contracts.ValidationError
		if errors.As(err, &validationErr) {
			if validationErr.Path != "" {
				lines = append(lines, line{"error_path", validationErr.Path})
			}
			lines = append(lines, line{"error_message", validationErr.Message})
		} else {
			lines = append(lines, line{"error_message", err.Error()})
		}
		writeLines(stderr, lines...)
		return exitInvalid
	}
	defer index.Close()

	output := []line{
		{"result", "ok"},
		{"command", "validate"},
		{"root", absRoot},
	}
	diagnostics := collectDiagnostics(index)
	if index.Resolution != nil {
		output = append(output,
			line{"selected_config_path", index.Resolution.SelectedConfigPath},
			line{"project_root", index.Resolution.ProjectRoot},
			line{"source_root", index.Resolution.SourceRoot},
			line{"source_mode", string(index.Resolution.SourceMode)},
			line{"source_ref", index.Resolution.SourceRef},
			line{"verification_posture", string(index.Resolution.VerificationPosture)},
			line{"diagnostic_count", fmt.Sprintf("%d", len(diagnostics))},
		)
		if index.Resolution.ResolvedCommit != "" {
			output = append(output, line{"resolved_commit", index.Resolution.ResolvedCommit})
		}
		if index.Resolution.VerifiedSignerIdentity != "" {
			output = append(output, line{"verified_signer_identity", index.Resolution.VerifiedSignerIdentity})
		}
		if index.Resolution.VerifiedSignerFingerprint != "" {
			output = append(output, line{"verified_signer_fingerprint", index.Resolution.VerifiedSignerFingerprint})
		}
	}
	for i, diagnostic := range diagnostics {
		prefix := fmt.Sprintf("diagnostic_%d", i+1)
		output = append(output,
			line{prefix + "_severity", string(diagnostic.Severity)},
			line{prefix + "_code", diagnostic.Code},
			line{prefix + "_message", diagnostic.Message},
		)
		if diagnostic.Path != "" {
			output = append(output, line{prefix + "_path", diagnostic.Path})
		}
		if diagnostic.Bundle != "" {
			output = append(output, line{prefix + "_bundle", diagnostic.Bundle})
		}
		if diagnostic.Aspect != "" {
			output = append(output, line{prefix + "_aspect", diagnostic.Aspect})
		}
		if diagnostic.Rule != "" {
			output = append(output, line{prefix + "_rule", diagnostic.Rule})
		}
		if diagnostic.Pattern != "" {
			output = append(output, line{prefix + "_pattern", diagnostic.Pattern})
		}
		if len(diagnostic.Matches) > 0 {
			output = append(output, line{prefix + "_matches", strings.Join(diagnostic.Matches, ",")})
		}
	}

	writeLines(stdout, output...)
	return exitOK
}

func runStatus(args []string, stdout, stderr io.Writer) int {
	request, err := parseStatusArgs(args)
	if err != nil {
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"command", "status"},
			line{"error_message", err.Error()},
			line{"usage", statusUsage},
		)
		return exitUsage
	}
	absRoot, validator, loaded, err := loadProjectForCLI(request.root, request.explicitRoot)
	if err != nil {
		writeLines(stderr,
			line{"result", "invalid"},
			line{"command", "status"},
			line{"root", absRootOrFallback(request.root, absRoot)},
			line{"error_message", err.Error()},
		)
		return exitInvalid
	}
	defer loaded.Close()
	summary, err := contracts.BuildProjectStatusSummary(validator, loaded)
	if err != nil {
		writeLines(stderr,
			line{"result", "invalid"},
			line{"command", "status"},
			line{"root", absRoot},
			line{"error_message", err.Error()},
		)
		return exitInvalid
	}
	output := []line{
		{"result", "ok"},
		{"command", "status"},
		{"root", absRoot},
		{"selected_config_path", summary.SelectedConfigPath},
		{"runecontext_version", summary.RuneContextVersion},
		{"assurance_tier", summary.AssuranceTier},
		{"active_count", fmt.Sprintf("%d", len(summary.Active))},
	}
	output = appendStatusEntries(output, "active", summary.Active)
	output = append(output, line{"closed_count", fmt.Sprintf("%d", len(summary.Closed))})
	output = appendStatusEntries(output, "closed", summary.Closed)
	output = append(output, line{"superseded_count", fmt.Sprintf("%d", len(summary.Superseded))})
	output = appendStatusEntries(output, "superseded", summary.Superseded)
	output = append(output, line{"bundle_count", fmt.Sprintf("%d", len(summary.BundleIDs))})
	for i, bundleID := range summary.BundleIDs {
		output = append(output, line{fmt.Sprintf("bundle_%d", i+1), bundleID})
	}
	writeLines(stdout, output...)
	return exitOK
}

func runChange(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"command", "change"},
			line{"error_message", "change subcommand is required"},
			line{"usage", changeUsage},
		)
		return exitUsage
	}
	switch args[0] {
	case "new":
		return runChangeNew(args[1:], stdout, stderr)
	case "shape":
		return runChangeShape(args[1:], stdout, stderr)
	case "close":
		return runChangeClose(args[1:], stdout, stderr)
	case "reallocate":
		return runChangeReallocate(args[1:], stdout, stderr)
	default:
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"command", "change"},
			line{"error_message", fmt.Sprintf("unknown change subcommand %q", args[0])},
			line{"usage", changeUsage},
		)
		return exitUsage
	}
}

func runChangeReallocate(args []string, stdout, stderr io.Writer) int {
	request, err := parseChangeReallocateArgs(args)
	if err != nil {
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"command", "change_reallocate"},
			line{"error_message", err.Error()},
			line{"usage", changeReallocateUsage},
		)
		return exitUsage
	}
	absRoot, validator, loaded, err := loadProjectForCLI(request.root, request.explicitRoot)
	if err != nil {
		writeLines(stderr,
			line{"result", "invalid"},
			line{"command", "change_reallocate"},
			line{"root", absRootOrFallback(request.root, absRoot)},
			line{"error_message", err.Error()},
		)
		return exitInvalid
	}
	defer loaded.Close()
	result, err := contracts.ReallocateChange(validator, loaded, request.changeID, contracts.ChangeReallocateOptions{})
	if err != nil {
		writeLines(stderr,
			line{"result", "invalid"},
			line{"command", "change_reallocate"},
			line{"root", absRoot},
			line{"error_message", err.Error()},
		)
		return exitInvalid
	}
	output := []line{
		{"result", "ok"},
		{"command", "change_reallocate"},
		{"root", absRoot},
		{"selected_config_path", loaded.Resolution.SelectedConfigPath},
		{"old_change_id", result.OldID},
		{"change_id", result.ID},
		{"old_change_path", result.OldChangePath},
		{"change_path", result.ChangePath},
		{"rewritten_reference_count", fmt.Sprintf("%d", result.RewrittenReferenceCount)},
	}
	output = appendWarnings(output, result.Warnings)
	output = appendChangedFiles(output, result.ChangedFiles)
	writeLines(stdout, output...)
	return exitOK
}

func runChangeNew(args []string, stdout, stderr io.Writer) int {
	request, err := parseChangeNewArgs(args)
	if err != nil {
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"command", "change_new"},
			line{"error_message", err.Error()},
			line{"usage", changeNewUsage},
		)
		return exitUsage
	}
	absRoot, validator, loaded, err := loadProjectForCLI(request.root, request.explicitRoot)
	if err != nil {
		writeLines(stderr,
			line{"result", "invalid"},
			line{"command", "change_new"},
			line{"root", absRootOrFallback(request.root, absRoot)},
			line{"error_message", err.Error()},
		)
		return exitInvalid
	}
	defer loaded.Close()
	result, err := contracts.CreateChange(validator, loaded, contracts.ChangeCreateOptions{
		Title:          request.title,
		Type:           request.changeType,
		Size:           request.size,
		Description:    request.description,
		ContextBundles: request.contextBundles,
		RequestedMode:  contracts.ChangeMode(request.mode),
	})
	if err != nil {
		writeLines(stderr,
			line{"result", "invalid"},
			line{"command", "change_new"},
			line{"root", absRoot},
			line{"error_message", err.Error()},
		)
		return exitInvalid
	}
	output := []line{
		{"result", "ok"},
		{"command", "change_new"},
		{"root", absRoot},
		{"selected_config_path", loaded.Resolution.SelectedConfigPath},
		{"change_id", result.ID},
		{"change_path", result.ChangePath},
		{"change_mode", string(result.Mode)},
		{"recommended_mode", string(result.RecommendedMode)},
		{"change_status", result.Status},
		{"context_bundle_count", fmt.Sprintf("%d", len(result.ContextBundles))},
	}
	output = appendStringItems(output, "context_bundle", result.ContextBundles)
	output = append(output, line{"applicable_standard_count", fmt.Sprintf("%d", len(result.ApplicableStandards))})
	output = appendStringItems(output, "applicable_standard", result.ApplicableStandards)
	output = append(output,
		line{"standards_refresh", result.StandardsRefreshAction},
		line{"review_diff_required", fmt.Sprintf("%t", result.ReviewDiffRequired)},
	)
	output = appendReasonsAndAssumptions(output, result.Reasons, result.Assumptions)
	output = appendChangedFiles(output, result.ChangedFiles)
	writeLines(stdout, output...)
	return exitOK
}

func runChangeShape(args []string, stdout, stderr io.Writer) int {
	request, err := parseChangeShapeArgs(args)
	if err != nil {
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"command", "change_shape"},
			line{"error_message", err.Error()},
			line{"usage", changeShapeUsage},
		)
		return exitUsage
	}
	absRoot, validator, loaded, err := loadProjectForCLI(request.root, request.explicitRoot)
	if err != nil {
		writeLines(stderr,
			line{"result", "invalid"},
			line{"command", "change_shape"},
			line{"root", absRootOrFallback(request.root, absRoot)},
			line{"error_message", err.Error()},
		)
		return exitInvalid
	}
	defer loaded.Close()
	result, err := contracts.ShapeChange(validator, loaded, request.changeID, contracts.ChangeShapeOptions{
		Design:       request.design,
		Verification: request.verification,
		Tasks:        request.tasks,
		References:   request.references,
	})
	if err != nil {
		writeLines(stderr,
			line{"result", "invalid"},
			line{"command", "change_shape"},
			line{"root", absRoot},
			line{"error_message", err.Error()},
		)
		return exitInvalid
	}
	output := []line{
		{"result", "ok"},
		{"command", "change_shape"},
		{"root", absRoot},
		{"selected_config_path", loaded.Resolution.SelectedConfigPath},
		{"change_id", result.ID},
		{"change_path", result.ChangePath},
		{"change_mode", string(result.Mode)},
		{"change_status", result.Status},
		{"applicable_standard_count", fmt.Sprintf("%d", len(result.ApplicableStandards))},
	}
	output = appendStringItems(output, "applicable_standard", result.ApplicableStandards)
	output = append(output, line{"added_standard_count", fmt.Sprintf("%d", len(result.AddedStandards))})
	output = appendStringItems(output, "added_standard", result.AddedStandards)
	output = append(output,
		line{"standards_refresh", result.StandardsRefreshAction},
		line{"review_diff_required", fmt.Sprintf("%t", result.ReviewDiffRequired)},
	)
	output = appendReasonsAndAssumptions(output, result.Reasons, result.Assumptions)
	output = appendChangedFiles(output, result.ChangedFiles)
	writeLines(stdout, output...)
	return exitOK
}

func runChangeClose(args []string, stdout, stderr io.Writer) int {
	request, err := parseChangeCloseArgs(args)
	if err != nil {
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"command", "change_close"},
			line{"error_message", err.Error()},
			line{"usage", changeCloseUsage},
		)
		return exitUsage
	}
	absRoot, validator, loaded, err := loadProjectForCLI(request.root, request.explicitRoot)
	if err != nil {
		writeLines(stderr,
			line{"result", "invalid"},
			line{"command", "change_close"},
			line{"root", absRootOrFallback(request.root, absRoot)},
			line{"error_message", err.Error()},
		)
		return exitInvalid
	}
	defer loaded.Close()
	result, err := contracts.CloseChange(validator, loaded, request.changeID, contracts.ChangeCloseOptions{
		VerificationStatus: request.verificationStatus,
		ClosedAt:           request.closedAt,
		SupersededBy:       request.supersededBy,
	})
	if err != nil {
		writeLines(stderr,
			line{"result", "invalid"},
			line{"command", "change_close"},
			line{"root", absRoot},
			line{"error_message", err.Error()},
		)
		return exitInvalid
	}
	output := []line{
		{"result", "ok"},
		{"command", "change_close"},
		{"root", absRoot},
		{"selected_config_path", loaded.Resolution.SelectedConfigPath},
		{"change_id", result.ID},
		{"change_path", result.ChangePath},
		{"change_mode", string(result.Mode)},
		{"change_status", result.Status},
	}
	if result.ClosedAt != "" {
		output = append(output, line{"closed_at", result.ClosedAt})
	}
	output = appendChangedFiles(output, result.ChangedFiles)
	writeLines(stdout, output...)
	return exitOK
}

type validateRequest struct {
	root         string
	explicitRoot bool
	gitTrust     contracts.GitTrustInputs
}

func parseValidateArgs(args []string) (validateRequest, error) {
	request := validateRequest{root: "."}
	var allowedSignersPath string
	positionals := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--ssh-allowed-signers":
			if i+1 >= len(args) {
				return validateRequest{}, fmt.Errorf("--ssh-allowed-signers requires a path")
			}
			i++
			allowedSignersPath = strings.TrimSpace(args[i])
			if allowedSignersPath == "" {
				return validateRequest{}, fmt.Errorf("--ssh-allowed-signers requires a path")
			}
		case strings.HasPrefix(arg, "--ssh-allowed-signers="):
			allowedSignersPath = strings.TrimSpace(strings.TrimPrefix(arg, "--ssh-allowed-signers="))
			if allowedSignersPath == "" {
				return validateRequest{}, fmt.Errorf("--ssh-allowed-signers requires a path")
			}
		case strings.HasPrefix(arg, "-"):
			return validateRequest{}, fmt.Errorf("unknown validate flag %q", arg)
		default:
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) > 1 {
		return validateRequest{}, fmt.Errorf("expected at most one path argument")
	}
	if len(positionals) == 1 {
		request.root = positionals[0]
		request.explicitRoot = true
	}
	if allowedSignersPath == "" {
		return request, nil
	}
	allowedSignersData, err := os.ReadFile(allowedSignersPath)
	if err != nil {
		return validateRequest{}, fmt.Errorf("read ssh allowed signers file %q: %w", allowedSignersPath, err)
	}
	verifier, err := contracts.NewSSHAllowedSignersVerifier(allowedSignersData)
	if err != nil {
		return validateRequest{}, fmt.Errorf("load ssh allowed signers file %q: %w", allowedSignersPath, err)
	}
	request.gitTrust.SignedTagVerifier = verifier
	return request, nil
}

type statusRequest struct {
	root         string
	explicitRoot bool
}

type changeNewRequest struct {
	root           string
	explicitRoot   bool
	title          string
	changeType     string
	size           string
	description    string
	mode           string
	contextBundles []string
}

type changeShapeRequest struct {
	root         string
	explicitRoot bool
	changeID     string
	design       string
	verification string
	tasks        []string
	references   []string
}

type changeCloseRequest struct {
	root               string
	explicitRoot       bool
	changeID           string
	verificationStatus string
	closedAt           time.Time
	supersededBy       []string
}

type changeReallocateRequest struct {
	root         string
	explicitRoot bool
	changeID     string
}

func parseStatusArgs(args []string) (statusRequest, error) {
	request := statusRequest{root: "."}
	positionals := make([]string, 0, 1)
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return statusRequest{}, fmt.Errorf("unknown status flag %q", arg)
		}
		positionals = append(positionals, arg)
	}
	if len(positionals) > 1 {
		return statusRequest{}, fmt.Errorf("expected at most one path argument")
	}
	if len(positionals) == 1 {
		request.root = positionals[0]
		request.explicitRoot = true
	}
	return request, nil
}

func parseChangeNewArgs(args []string) (changeNewRequest, error) {
	request := changeNewRequest{root: "."}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--title":
			value, next, err := requireFlagValue(args, i, "--title")
			if err != nil {
				return changeNewRequest{}, err
			}
			request.title = value
			i = next
		case strings.HasPrefix(arg, "--title="):
			request.title = strings.TrimSpace(strings.TrimPrefix(arg, "--title="))
		case arg == "--type":
			value, next, err := requireFlagValue(args, i, "--type")
			if err != nil {
				return changeNewRequest{}, err
			}
			request.changeType = value
			i = next
		case strings.HasPrefix(arg, "--type="):
			request.changeType = strings.TrimSpace(strings.TrimPrefix(arg, "--type="))
		case arg == "--size":
			value, next, err := requireFlagValue(args, i, "--size")
			if err != nil {
				return changeNewRequest{}, err
			}
			request.size = value
			i = next
		case strings.HasPrefix(arg, "--size="):
			request.size = strings.TrimSpace(strings.TrimPrefix(arg, "--size="))
		case arg == "--description":
			value, next, err := requireFlagValue(args, i, "--description")
			if err != nil {
				return changeNewRequest{}, err
			}
			request.description = value
			i = next
		case strings.HasPrefix(arg, "--description="):
			request.description = strings.TrimSpace(strings.TrimPrefix(arg, "--description="))
		case arg == "--shape":
			value, next, err := requireFlagValue(args, i, "--shape")
			if err != nil {
				return changeNewRequest{}, err
			}
			request.mode = value
			i = next
		case strings.HasPrefix(arg, "--shape="):
			request.mode = strings.TrimSpace(strings.TrimPrefix(arg, "--shape="))
		case arg == "--bundle":
			value, next, err := requireFlagValue(args, i, "--bundle")
			if err != nil {
				return changeNewRequest{}, err
			}
			request.contextBundles = append(request.contextBundles, value)
			i = next
		case strings.HasPrefix(arg, "--bundle="):
			request.contextBundles = append(request.contextBundles, strings.TrimSpace(strings.TrimPrefix(arg, "--bundle=")))
		case arg == "--path":
			value, next, err := requireFlagValue(args, i, "--path")
			if err != nil {
				return changeNewRequest{}, err
			}
			request.root = value
			request.explicitRoot = true
			i = next
		case strings.HasPrefix(arg, "--path="):
			request.root = strings.TrimSpace(strings.TrimPrefix(arg, "--path="))
			request.explicitRoot = true
		case strings.HasPrefix(arg, "-"):
			return changeNewRequest{}, fmt.Errorf("unknown change new flag %q", arg)
		default:
			return changeNewRequest{}, fmt.Errorf("unexpected positional argument %q", arg)
		}
	}
	if strings.TrimSpace(request.title) == "" {
		return changeNewRequest{}, fmt.Errorf("--title is required")
	}
	if strings.TrimSpace(request.changeType) == "" {
		return changeNewRequest{}, fmt.Errorf("--type is required")
	}
	if request.mode != "" && request.mode != string(contracts.ChangeModeMinimum) && request.mode != string(contracts.ChangeModeFull) {
		return changeNewRequest{}, fmt.Errorf("--shape must be %q or %q", contracts.ChangeModeMinimum, contracts.ChangeModeFull)
	}
	return request, nil
}

func parseChangeShapeArgs(args []string) (changeShapeRequest, error) {
	request := changeShapeRequest{root: "."}
	positionals := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--design":
			value, next, err := requireFlagValue(args, i, "--design")
			if err != nil {
				return changeShapeRequest{}, err
			}
			request.design = value
			i = next
		case strings.HasPrefix(arg, "--design="):
			request.design = strings.TrimSpace(strings.TrimPrefix(arg, "--design="))
		case arg == "--verification":
			value, next, err := requireFlagValue(args, i, "--verification")
			if err != nil {
				return changeShapeRequest{}, err
			}
			request.verification = value
			i = next
		case strings.HasPrefix(arg, "--verification="):
			request.verification = strings.TrimSpace(strings.TrimPrefix(arg, "--verification="))
		case arg == "--task":
			value, next, err := requireFlagValue(args, i, "--task")
			if err != nil {
				return changeShapeRequest{}, err
			}
			request.tasks = append(request.tasks, value)
			i = next
		case strings.HasPrefix(arg, "--task="):
			request.tasks = append(request.tasks, strings.TrimSpace(strings.TrimPrefix(arg, "--task=")))
		case arg == "--reference":
			value, next, err := requireFlagValue(args, i, "--reference")
			if err != nil {
				return changeShapeRequest{}, err
			}
			request.references = append(request.references, value)
			i = next
		case strings.HasPrefix(arg, "--reference="):
			request.references = append(request.references, strings.TrimSpace(strings.TrimPrefix(arg, "--reference=")))
		case arg == "--path":
			value, next, err := requireFlagValue(args, i, "--path")
			if err != nil {
				return changeShapeRequest{}, err
			}
			request.root = value
			request.explicitRoot = true
			i = next
		case strings.HasPrefix(arg, "--path="):
			request.root = strings.TrimSpace(strings.TrimPrefix(arg, "--path="))
			request.explicitRoot = true
		case strings.HasPrefix(arg, "-"):
			return changeShapeRequest{}, fmt.Errorf("unknown change shape flag %q", arg)
		default:
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) != 1 {
		return changeShapeRequest{}, fmt.Errorf("change shape requires exactly one change ID")
	}
	request.changeID = positionals[0]
	return request, nil
}

func parseChangeCloseArgs(args []string) (changeCloseRequest, error) {
	request := changeCloseRequest{root: "."}
	positionals := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--verification-status":
			value, next, err := requireFlagValue(args, i, "--verification-status")
			if err != nil {
				return changeCloseRequest{}, err
			}
			request.verificationStatus = value
			i = next
		case strings.HasPrefix(arg, "--verification-status="):
			request.verificationStatus = strings.TrimSpace(strings.TrimPrefix(arg, "--verification-status="))
		case arg == "--superseded-by":
			value, next, err := requireFlagValue(args, i, "--superseded-by")
			if err != nil {
				return changeCloseRequest{}, err
			}
			request.supersededBy = append(request.supersededBy, value)
			i = next
		case strings.HasPrefix(arg, "--superseded-by="):
			request.supersededBy = append(request.supersededBy, strings.TrimSpace(strings.TrimPrefix(arg, "--superseded-by=")))
		case arg == "--closed-at":
			value, next, err := requireFlagValue(args, i, "--closed-at")
			if err != nil {
				return changeCloseRequest{}, err
			}
			parsed, err := time.Parse("2006-01-02", value)
			if err != nil {
				return changeCloseRequest{}, fmt.Errorf("--closed-at must use YYYY-MM-DD")
			}
			request.closedAt = parsed
			i = next
		case strings.HasPrefix(arg, "--closed-at="):
			parsed, err := time.Parse("2006-01-02", strings.TrimSpace(strings.TrimPrefix(arg, "--closed-at=")))
			if err != nil {
				return changeCloseRequest{}, fmt.Errorf("--closed-at must use YYYY-MM-DD")
			}
			request.closedAt = parsed
		case arg == "--path":
			value, next, err := requireFlagValue(args, i, "--path")
			if err != nil {
				return changeCloseRequest{}, err
			}
			request.root = value
			request.explicitRoot = true
			i = next
		case strings.HasPrefix(arg, "--path="):
			request.root = strings.TrimSpace(strings.TrimPrefix(arg, "--path="))
			request.explicitRoot = true
		case strings.HasPrefix(arg, "-"):
			return changeCloseRequest{}, fmt.Errorf("unknown change close flag %q", arg)
		default:
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) != 1 {
		return changeCloseRequest{}, fmt.Errorf("change close requires exactly one change ID")
	}
	request.changeID = positionals[0]
	return request, nil
}

func parseChangeReallocateArgs(args []string) (changeReallocateRequest, error) {
	request := changeReallocateRequest{root: "."}
	positionals := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--path":
			value, next, err := requireFlagValue(args, i, "--path")
			if err != nil {
				return changeReallocateRequest{}, err
			}
			request.root = value
			request.explicitRoot = true
			i = next
		case strings.HasPrefix(arg, "--path="):
			request.root = strings.TrimSpace(strings.TrimPrefix(arg, "--path="))
			request.explicitRoot = true
		case strings.HasPrefix(arg, "-"):
			return changeReallocateRequest{}, fmt.Errorf("unknown change reallocate flag %q", arg)
		default:
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) != 1 {
		return changeReallocateRequest{}, fmt.Errorf("change reallocate requires exactly one change ID")
	}
	request.changeID = positionals[0]
	return request, nil
}

func requireFlagValue(args []string, index int, flag string) (string, int, error) {
	if index+1 >= len(args) {
		return "", index, fmt.Errorf("%s requires a value", flag)
	}
	value := strings.TrimSpace(args[index+1])
	if value == "" {
		return "", index, fmt.Errorf("%s requires a value", flag)
	}
	return value, index + 1, nil
}

func loadProjectForCLI(root string, explicitRoot bool) (string, *contracts.Validator, *contracts.LoadedProject, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", nil, nil, err
	}
	schemaRoot, err := locateSchemaRoot()
	if err != nil {
		return absRoot, nil, nil, err
	}
	validator := contracts.NewValidator(schemaRoot)
	options := contracts.ResolveOptions{
		ConfigDiscovery: contracts.ConfigDiscoveryNearestAncestor,
		ExecutionMode:   contracts.ExecutionModeLocal,
	}
	if explicitRoot {
		options.ConfigDiscovery = contracts.ConfigDiscoveryExplicitRoot
	}
	loaded, err := validator.LoadProject(absRoot, options)
	if err != nil {
		return absRoot, nil, nil, err
	}
	return absRoot, validator, loaded, nil
}

func absRootOrFallback(root, absRoot string) string {
	if absRoot != "" {
		return absRoot
	}
	if value, err := filepath.Abs(root); err == nil {
		return value
	}
	return root
}

func appendStatusEntries(lines []line, prefix string, entries []contracts.ChangeStatusEntry) []line {
	for i, entry := range entries {
		index := i + 1
		lines = append(lines,
			line{fmt.Sprintf("%s_%d_id", prefix, index), entry.ID},
			line{fmt.Sprintf("%s_%d_title", prefix, index), entry.Title},
			line{fmt.Sprintf("%s_%d_status", prefix, index), entry.Status},
			line{fmt.Sprintf("%s_%d_type", prefix, index), entry.Type},
			line{fmt.Sprintf("%s_%d_size", prefix, index), entry.Size},
			line{fmt.Sprintf("%s_%d_path", prefix, index), entry.Path},
		)
	}
	return lines
}

func appendStringItems(lines []line, prefix string, items []string) []line {
	for i, item := range items {
		lines = append(lines, line{fmt.Sprintf("%s_%d", prefix, i+1), item})
	}
	return lines
}

func appendReasonsAndAssumptions(lines []line, reasons, assumptions []string) []line {
	lines = append(lines, line{"reason_count", fmt.Sprintf("%d", len(reasons))})
	for i, reason := range reasons {
		lines = append(lines, line{fmt.Sprintf("reason_%d", i+1), reason})
	}
	lines = append(lines, line{"assumption_count", fmt.Sprintf("%d", len(assumptions))})
	for i, assumption := range assumptions {
		lines = append(lines, line{fmt.Sprintf("assumption_%d", i+1), assumption})
	}
	return lines
}

func appendWarnings(lines []line, warnings []string) []line {
	lines = append(lines, line{"warning_count", fmt.Sprintf("%d", len(warnings))})
	for i, warning := range warnings {
		lines = append(lines, line{fmt.Sprintf("warning_%d", i+1), warning})
	}
	return lines
}

func appendChangedFiles(lines []line, changed []contracts.FileMutation) []line {
	lines = append(lines, line{"changed_file_count", fmt.Sprintf("%d", len(changed))})
	for i, file := range changed {
		prefix := fmt.Sprintf("changed_file_%d", i+1)
		lines = append(lines,
			line{prefix + "_path", file.Path},
			line{prefix + "_action", file.Action},
		)
	}
	return lines
}

type line struct {
	key   string
	value string
}

func locateSchemaRoot() (string, error) {
	starts := make([]string, 0, 2)
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
	seen := map[string]struct{}{}
	for _, start := range starts {
		if start == "" {
			continue
		}
		clean := filepath.Clean(start)
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		if root, ok := findSchemaRoot(clean); ok {
			return root, nil
		}
	}
	return "", fmt.Errorf("could not locate RuneContext schemas from the current working directory or executable location")
}

func findSchemaRoot(start string) (string, bool) {
	current := start
	if info, err := os.Stat(current); err == nil && !info.IsDir() {
		current = filepath.Dir(current)
	}
	for {
		if isSchemaDir(current) {
			return current, true
		}
		candidate := filepath.Join(current, "schemas")
		if isSchemaDir(candidate) {
			return candidate, true
		}
		next := filepath.Dir(current)
		if next == current {
			return "", false
		}
		current = next
	}
}

func isSchemaDir(path string) bool {
	for _, name := range []string{"runecontext.schema.json", "bundle.schema.json", "change-status.schema.json", "context-pack.schema.json"} {
		if _, err := os.Stat(filepath.Join(path, name)); err != nil {
			return false
		}
	}
	return true
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "RuneContext CLI")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  runectx help")
	fmt.Fprintln(w, "  "+statusUsage)
	fmt.Fprintln(w, "  "+changeNewUsage)
	fmt.Fprintln(w, "  "+changeShapeUsage)
	fmt.Fprintln(w, "  "+changeCloseUsage)
	fmt.Fprintln(w, "  "+changeReallocateUsage)
	fmt.Fprintln(w, "  "+validateUsage)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  help       Show CLI usage")
	fmt.Fprintln(w, "  status     Report active, closed, and superseded changes")
	fmt.Fprintln(w, "  change     Create, shape, close, and reallocate changes")
	fmt.Fprintln(w, "  validate   Validate RuneContext contracts for a project root")
}

func writeLines(w io.Writer, lines ...line) {
	for _, entry := range lines {
		fmt.Fprintf(w, "%s=%s\n", entry.key, sanitizeValue(entry.value))
	}
}

func sanitizeValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, "\t", "\\t")
	value = strings.ReplaceAll(value, "\x00", "\\0")
	value = strings.ReplaceAll(value, "=", "\\=")
	return value
}

type emittedDiagnostic struct {
	Severity contracts.DiagnosticSeverity
	Code     string
	Message  string
	Path     string
	Bundle   string
	Aspect   string
	Rule     string
	Pattern  string
	Matches  []string
}

func collectDiagnostics(index *contracts.ProjectIndex) []emittedDiagnostic {
	if index == nil {
		return nil
	}
	items := make([]emittedDiagnostic, 0)
	if index.Resolution != nil {
		for _, diagnostic := range index.Resolution.Diagnostics {
			items = append(items, emittedDiagnostic{Severity: diagnostic.Severity, Code: diagnostic.Code, Message: diagnostic.Message})
		}
	}
	for _, diagnostic := range index.Diagnostics {
		items = append(items, emittedDiagnostic{Severity: diagnostic.Severity, Code: diagnostic.Code, Message: diagnostic.Message, Path: diagnostic.Path})
	}
	if index.Bundles != nil {
		for _, diagnostic := range index.Bundles.Diagnostics() {
			items = append(items, emittedDiagnostic{Severity: diagnostic.Severity, Code: diagnostic.Code, Message: diagnostic.Message, Bundle: diagnostic.Bundle, Aspect: string(diagnostic.Aspect), Rule: string(diagnostic.Rule), Pattern: diagnostic.Pattern, Matches: append([]string(nil), diagnostic.Matches...)})
		}
	}
	return dedupeDiagnostics(items)
}

func dedupeDiagnostics(items []emittedDiagnostic) []emittedDiagnostic {
	if len(items) == 0 {
		return nil
	}
	result := make([]emittedDiagnostic, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		key := strings.Join([]string{string(item.Severity), item.Code, item.Message, item.Path, item.Bundle, item.Aspect, item.Rule, item.Pattern, strings.Join(item.Matches, ",")}, "\x1f")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}

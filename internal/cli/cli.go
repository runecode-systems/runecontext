package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

const (
	exitOK      = 0
	exitInvalid = 1
	exitUsage   = 2
)

const validateUsage = "runectx validate [--ssh-allowed-signers PATH] [path]"

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return exitOK
	}

	switch args[0] {
	case "validate":
		return runValidate(args[1:], stdout, stderr)
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
	if root != "." {
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

type validateRequest struct {
	root     string
	gitTrust contracts.GitTrustInputs
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
	fmt.Fprintln(w, "  "+validateUsage)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  help       Show CLI usage")
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

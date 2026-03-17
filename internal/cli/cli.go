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
			line{"usage", "runectx validate [path]"},
		)
		return exitUsage
	}
}

func runValidate(args []string, stdout, stderr io.Writer) int {
	if len(args) > 1 {
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"command", "validate"},
			line{"error_message", "expected at most one path argument"},
			line{"usage", "runectx validate [path]"},
		)
		return exitUsage
	}

	root := "."
	resolveOptions := contracts.ResolveOptions{
		ConfigDiscovery: contracts.ConfigDiscoveryNearestAncestor,
		ExecutionMode:   contracts.ExecutionModeLocal,
	}
	if len(args) == 1 {
		root = args[0]
		resolveOptions.ConfigDiscovery = contracts.ConfigDiscoveryExplicitRoot
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"command", "validate"},
			line{"error_message", fmt.Sprintf("failed to resolve path %q: %v", root, err)},
			line{"usage", "runectx validate [path]"},
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
	if index.Resolution != nil {
		output = append(output,
			line{"selected_config_path", index.Resolution.SelectedConfigPath},
			line{"project_root", index.Resolution.ProjectRoot},
			line{"source_root", index.Resolution.SourceRoot},
			line{"source_mode", string(index.Resolution.SourceMode)},
			line{"source_ref", index.Resolution.SourceRef},
			line{"verification_posture", string(index.Resolution.VerificationPosture)},
			line{"diagnostic_count", fmt.Sprintf("%d", len(index.Resolution.Diagnostics))},
		)
		if index.Resolution.ResolvedCommit != "" {
			output = append(output, line{"resolved_commit", index.Resolution.ResolvedCommit})
		}
		for i, diagnostic := range index.Resolution.Diagnostics {
			prefix := fmt.Sprintf("diagnostic_%d", i+1)
			output = append(output,
				line{prefix + "_severity", string(diagnostic.Severity)},
				line{prefix + "_code", diagnostic.Code},
				line{prefix + "_message", diagnostic.Message},
			)
		}
	}

	writeLines(stdout, output...)
	return exitOK
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
	fmt.Fprintln(w, "  runectx validate [path]")
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
	return value
}

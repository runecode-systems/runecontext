package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/runecode-ai/runecontext/internal/contracts"
)

const (
	exitOK      = 0
	exitInvalid = 1
	exitUsage   = 2
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeLines(stderr,
			line{"result", "usage_error"},
			line{"error_message", "missing command"},
			line{"usage", "runectx validate [path]"},
		)
		return exitUsage
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
	if len(args) == 1 {
		root = args[0]
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

	repo := repoRoot(absRoot)
	validator := contracts.NewValidator(filepath.Join(repo, "schemas"))
	if _, err := validator.ValidateProject(absRoot); err != nil {
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

	writeLines(stdout,
		line{"result", "ok"},
		line{"command", "validate"},
		line{"root", absRoot},
	)
	return exitOK
}

type line struct {
	key   string
	value string
}

func repoRoot(start string) string {
	current := start
	if info, err := os.Stat(current); err == nil && !info.IsDir() {
		current = filepath.Dir(current)
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "schemas")); err == nil {
			if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
				return current
			}
		}
		next := filepath.Dir(current)
		if next == current {
			return start
		}
		current = next
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "RuneContext CLI")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  runectx validate [path]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
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

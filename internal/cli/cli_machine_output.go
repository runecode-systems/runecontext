package cli

import (
	"encoding/json"
	"io"
)

type failureClass string

const (
	failureClassNone    failureClass = "none"
	failureClassInvalid failureClass = "invalid"
	failureClassUsage   failureClass = "usage"
)

type machineEnvelope struct {
	SchemaVersion int               `json:"schema_version"`
	Result        string            `json:"result"`
	Command       string            `json:"command"`
	ExitCode      int               `json:"exit_code"`
	FailureClass  failureClass      `json:"failure_class"`
	Data          map[string]string `json:"data"`
}

func emitOutput(w io.Writer, options machineOptions, lines []line, exitCode int, class failureClass) {
	if !options.jsonOutput {
		writeLines(w, lines...)
		return
	}
	envelope := machineEnvelope{
		SchemaVersion: 1,
		Result:        lineValue(lines, "result"),
		Command:       lineValue(lines, "command"),
		ExitCode:      exitCode,
		FailureClass:  class,
		Data:          linesToMap(lines),
	}
	if duplicateKey, ok := firstDuplicateKey(lines); ok {
		envelope.Result = "invalid"
		envelope.FailureClass = failureClassInvalid
		envelope.Data = map[string]string{"error_message": "duplicate output keys", "duplicate_key": duplicateKey}
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		fallback := machineEnvelope{
			SchemaVersion: 1,
			Result:        "invalid",
			Command:       lineValue(lines, "command"),
			ExitCode:      exitCode,
			FailureClass:  failureClassInvalid,
			Data:          map[string]string{"error_message": "failed to encode --json output"},
		}
		fallbackData, _ := json.Marshal(fallback)
		w.Write(fallbackData)
		w.Write([]byte{'\n'})
		return
	}
	w.Write(data)
	w.Write([]byte{'\n'})
}
func linesToMap(lines []line) map[string]string {
	result := make(map[string]string, len(lines))
	for _, item := range lines {
		result[item.key] = item.value
	}
	return result
}

func lineValue(lines []line, key string) string {
	for _, item := range lines {
		if item.key == key {
			return item.value
		}
	}
	return ""
}

func appendMachineOptionLines(lines []line, options machineOptions) []line {
	lines = append(lines,
		line{"non_interactive", boolString(options.nonInteractive)},
		line{"dry_run", boolString(options.dryRun)},
		line{"explain", boolString(options.explain)},
	)
	return lines
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func assertUniqueKeys(lines []line) bool {
	_, ok := firstDuplicateKey(lines)
	return !ok
}

func firstDuplicateKey(lines []line) (string, bool) {
	seen := make(map[string]struct{}, len(lines))
	for _, item := range lines {
		if _, ok := seen[item.key]; ok {
			return item.key, true
		}
		seen[item.key] = struct{}{}
	}
	return "", false
}

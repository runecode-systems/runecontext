package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func parseCLIKeyValueOutput(t *testing.T, output string) map[string]string {
	t.Helper()
	fields := map[string]string{}
	foundKeyValue := false
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, "=") {
			t.Fatalf("expected key=value output, got %q", line)
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			t.Fatalf("malformed CLI output line: %q", line)
		}
		foundKeyValue = true
		fields[parts[0]] = unsanitizeCLIValue(parts[1])
	}
	if !foundKeyValue {
		t.Fatalf("expected at least one key=value line in output: %q", output)
	}
	return fields
}

func parseCLIJSONEnvelopeData(t *testing.T, payload []byte) map[string]string {
	t.Helper()
	var envelope struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("expected JSON output, got err=%v payload=%q", err, string(payload))
	}
	if envelope.Data == nil {
		t.Fatalf("expected JSON envelope data, got payload=%q", string(payload))
	}
	return envelope.Data
}

func unsanitizeCLIValue(value string) string {
	var builder strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] != '\\' || i+1 >= len(value) {
			builder.WriteByte(value[i])
			continue
		}
		i++
		switch value[i] {
		case '\\':
			builder.WriteByte('\\')
		case 'n':
			builder.WriteByte('\n')
		case 'r':
			builder.WriteByte('\r')
		case 't':
			builder.WriteByte('\t')
		case '0':
			builder.WriteByte('\x00')
		case '=':
			builder.WriteByte('=')
		default:
			builder.WriteByte('\\')
			builder.WriteByte(value[i])
		}
	}
	return builder.String()
}

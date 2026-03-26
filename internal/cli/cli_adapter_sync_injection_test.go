package cli

import "testing"

func TestSupportsShellInjectionMatrix(t *testing.T) {
	tests := []struct {
		tool string
		want bool
	}{
		{tool: "opencode", want: true},
		{tool: "claude-code", want: true},
		{tool: "codex", want: false},
		{tool: "generic", want: false},
	}
	for _, tc := range tests {
		if got := supportsShellInjection(tc.tool); got != tc.want {
			t.Fatalf("supportsShellInjection(%q)=%v, want %v", tc.tool, got, tc.want)
		}
	}
}

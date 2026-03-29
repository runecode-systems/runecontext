package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func appendStatusExplainHuman(builder *strings.Builder, loaded *contracts.LoadedProject, summary *contracts.ProjectStatusSummary, options statusRenderOptions) {
	lines := appendStatusExplainLines(nil, loaded, summary)
	if len(lines) == 0 {
		return
	}
	builder.WriteString(styleStatusText("Explain", ansiBold, options.color))
	builder.WriteString("\n")
	for _, item := range lines {
		builder.WriteString(fmt.Sprintf("- %s: %s\n", sanitizeStatusText(item.key), sanitizeStatusText(item.value)))
	}
	builder.WriteString("\n")
}

func sanitizeStatusText(value string) string {
	if value == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(value))
	for i := 0; i < len(value); {
		if value[i] == '\x1b' {
			i = consumeANSISequence(value, i)
			continue
		}
		r, size := utf8.DecodeRuneInString(value[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		if unicode.IsControl(r) {
			i += size
			continue
		}
		builder.WriteRune(r)
		i += size
	}
	return builder.String()
}

func consumeANSISequence(value string, start int) int {
	i := start + 1
	if i >= len(value) || value[i] != '[' {
		return i
	}
	i++
	for i < len(value) {
		if value[i] >= 0x40 && value[i] <= 0x7e {
			return i + 1
		}
		i++
	}
	return i
}

func shouldUseStatusColor(w io.Writer) bool {
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return false
	}
	if term := strings.TrimSpace(strings.ToLower(os.Getenv("TERM"))); term == "" || term == "dumb" {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

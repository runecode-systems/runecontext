package cli

import (
	"fmt"
	"io"
)

func runCompletion(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeCommandUsageError(stderr, "completion", completionUsage, fmt.Errorf("completion shell is required"))
		return exitUsage
	}
	if isHelpToken(args[0]) {
		if len(args) != 1 {
			writeCommandUsageError(stderr, "completion", completionUsage, fmt.Errorf("help does not accept additional arguments"))
			return exitUsage
		}
		writeLines(stdout,
			line{"result", "ok"},
			line{"command", "completion"},
			line{"usage", completionUsage},
		)
		return exitOK
	}
	if len(args) != 1 {
		writeCommandUsageError(stderr, "completion", completionUsage, fmt.Errorf("completion expects exactly one shell argument"))
		return exitUsage
	}
	script, err := generateCompletionScript(CommandMetadataRegistry(), args[0])
	if err != nil {
		writeCommandUsageError(stderr, "completion", completionUsage, err)
		return exitUsage
	}
	if _, err := io.WriteString(stdout, script); err != nil {
		writeCommandInvalid(stderr, "completion", "", err)
		return exitInvalid
	}
	return exitOK
}

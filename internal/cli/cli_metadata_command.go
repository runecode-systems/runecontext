package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

func runMetadata(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && isHelpToken(args[0]) {
		if len(args) != 1 {
			writeCommandUsageError(stderr, "metadata", metadataUsage, fmt.Errorf("help does not accept additional arguments"))
			return exitUsage
		}
		writeLines(stdout,
			line{"result", "ok"},
			line{"command", "metadata"},
			line{"usage", metadataUsage},
		)
		return exitOK
	}
	if len(args) > 0 {
		err := fmt.Errorf("metadata does not accept positional arguments")
		if len(args[0]) > 0 && args[0][0] == '-' {
			err = fmt.Errorf("unknown metadata flag %q", args[0])
		}
		writeCommandUsageError(stderr, "metadata", metadataUsage, err)
		return exitUsage
	}

	descriptor := buildCapabilityDescriptor()
	if err := validateCapabilityDescriptorSchema(descriptor); err != nil {
		writeCommandInvalid(stderr, "metadata", "", err)
		return exitInvalid
	}

	payload, err := json.MarshalIndent(descriptor, "", "  ")
	if err != nil {
		writeCommandInvalid(stderr, "metadata", "", err)
		return exitInvalid
	}
	if _, err := io.WriteString(stdout, string(payload)+"\n"); err != nil {
		writeCommandInvalid(stderr, "metadata", "", err)
		return exitInvalid
	}
	return exitOK
}

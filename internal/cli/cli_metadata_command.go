package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

func runMetadata(args []string, stdout, stderr io.Writer) int {
	if code, handled := handleMetadataArgs(args, stdout, stderr); handled {
		return code
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

func handleMetadataArgs(args []string, stdout, stderr io.Writer) (int, bool) {
	if len(args) == 0 {
		return 0, false
	}
	if isHelpToken(args[0]) {
		if len(args) != 1 {
			writeCommandUsageError(stderr, "metadata", metadataUsage, fmt.Errorf("help does not accept additional arguments"))
			return exitUsage, true
		}
		writeLines(stdout,
			line{"result", "ok"},
			line{"command", "metadata"},
			line{"usage", metadataUsage},
		)
		return exitOK, true
	}
	err := fmt.Errorf("metadata does not accept positional arguments")
	if args[0] != "" && args[0][0] == '-' {
		err = fmt.Errorf("unknown metadata flag %q", args[0])
	}
	writeCommandUsageError(stderr, "metadata", metadataUsage, err)
	return exitUsage, true
}

package cli

import (
	"fmt"
	"io"
)

const (
	exitOK      = 0
	exitInvalid = 1
	exitUsage   = 2
)

const (
	validateUsage         = "runectx validate [--ssh-allowed-signers PATH] [path]"
	statusUsage           = "runectx status [path]"
	changeUsage           = "runectx change <new|shape|close|reallocate> ..."
	changeNewUsage        = "runectx change new --title TITLE --type TYPE [--size SIZE] [--bundle ID] [--shape minimum|full] [--description TEXT] [--path PATH]"
	changeShapeUsage      = "runectx change shape CHANGE_ID [--design TEXT] [--verification TEXT] [--task TEXT] [--reference TEXT] [--path PATH]"
	changeCloseUsage      = "runectx change close CHANGE_ID [--verification-status STATUS] [--superseded-by ID] [--closed-at YYYY-MM-DD] [--path PATH]"
	changeReallocateUsage = "runectx change reallocate CHANGE_ID [--path PATH]"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return exitOK
	}

	switch args[0] {
	case "validate":
		return runValidate(args[1:], stdout, stderr)
	case "status":
		return runStatus(args[1:], stdout, stderr)
	case "change":
		return runChange(args[1:], stdout, stderr)
	case "help", "--help", "-h":
		printUsage(stdout)
		return exitOK
	default:
		writeCommandUsageError(stderr, args[0], validateUsage, fmt.Errorf("unknown command %q", args[0]))
		return exitUsage
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "RuneContext CLI")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  runectx help")
	fmt.Fprintln(w, "  "+statusUsage)
	fmt.Fprintln(w, "  "+changeNewUsage)
	fmt.Fprintln(w, "  "+changeShapeUsage)
	fmt.Fprintln(w, "  "+changeCloseUsage)
	fmt.Fprintln(w, "  "+changeReallocateUsage)
	fmt.Fprintln(w, "  "+validateUsage)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  help       Show CLI usage")
	fmt.Fprintln(w, "  status     Report active, closed, and superseded changes")
	fmt.Fprintln(w, "  change     Create, shape, close, and reallocate changes")
	fmt.Fprintln(w, "  validate   Validate RuneContext contracts for a project root")
}

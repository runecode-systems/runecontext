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
	validateUsage         = "runectx validate [--json] [--non-interactive] [--explain] [--ssh-allowed-signers PATH] [path]"
	statusUsage           = "runectx status [--json] [--non-interactive] [--explain] [path]"
	changeUsage           = "runectx change [--json] [--non-interactive] [--dry-run] [--explain] <new|shape|close|reallocate> ..."
	bundleUsage           = "runectx bundle [--json] [--non-interactive] [--explain] <resolve>"
	bundleResolveUsage    = "runectx bundle resolve [--json] [--non-interactive] [--explain] [--path PATH] <bundle-id>..."
	doctorUsage           = "runectx doctor [--json] [--non-interactive] [--explain] [--path PATH] [path]"
	changeNewUsage        = "runectx change new [--json] [--non-interactive] [--dry-run] [--explain] --title TITLE --type TYPE [--size SIZE] [--bundle ID] [--shape minimum|full] [--description TEXT] [--path PATH]"
	changeShapeUsage      = "runectx change shape [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--design TEXT] [--verification TEXT] [--task TEXT] [--reference TEXT] [--path PATH]"
	changeCloseUsage      = "runectx change close [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--verification-status STATUS] [--superseded-by ID] [--closed-at YYYY-MM-DD] [--path PATH]"
	changeReallocateUsage = "runectx change reallocate [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--path PATH]"
	initUsage             = "runectx init [--json] [--non-interactive] [--dry-run] [--explain] [--mode embedded|linked] [--seed-bundle NAME] [--path PATH]"
	promoteUsage          = "runectx promote [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--accept | --complete] [--target TYPE:PATH (summary auto-filled per target type)] [--path PATH]"
	standardUsage         = "runectx standard [--json] [--non-interactive] [--explain] <discover>"
	standardDiscoverUsage = "runectx standard discover [--json] [--non-interactive] [--explain] [--path PATH] [--change CHANGE_ID] [--confirm-handoff] [--target TYPE:PATH]"
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
	case "bundle":
		return runBundle(args[1:], stdout, stderr)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "promote":
		return runPromote(args[1:], stdout, stderr)
	case "standard":
		return runStandard(args[1:], stdout, stderr)
	case "help", "--help", "-h":
		printUsage(stdout)
		return exitOK
	default:
		writeCommandUsageError(stderr, args[0], "runectx help", fmt.Errorf("unknown command %q", args[0]))
		return exitUsage
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "RuneContext CLI")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  help       Show CLI usage")
	fmt.Fprintln(w, "  status     Report active, closed, and superseded changes")
	fmt.Fprintln(w, "  change     Create, shape, close, and reallocate changes")
	fmt.Fprintln(w, "  bundle     Resolve context bundles")
	fmt.Fprintln(w, "  validate   Validate RuneContext contracts for a project root")
	fmt.Fprintln(w, "  doctor     Run environment and resolution diagnostics")
	fmt.Fprintln(w, "  init       Scaffold a RuneContext project")
	fmt.Fprintln(w, "  promote    Explicitly advance promotion assessment state (summary auto-filled for --target entries)")
	fmt.Fprintln(w, "  standard   Discover advisory standards candidates for promotion handoff")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  runectx help")
	fmt.Fprintln(w, "  "+statusUsage)
	fmt.Fprintln(w, "  "+changeNewUsage)
	fmt.Fprintln(w, "  "+changeShapeUsage)
	fmt.Fprintln(w, "  "+changeCloseUsage)
	fmt.Fprintln(w, "  "+changeReallocateUsage)
	fmt.Fprintln(w, "  "+validateUsage)
	fmt.Fprintln(w, "  "+bundleResolveUsage)
	fmt.Fprintln(w, "  "+doctorUsage)
	fmt.Fprintln(w, "  "+initUsage)
	fmt.Fprintln(w, "  "+promoteUsage)
	fmt.Fprintln(w, "  "+standardDiscoverUsage)
}

package contracts

import (
	"context"
	"fmt"
	"os/exec"
)

func runGit(args ...string) error {
	result := runGitCaptured(args, gitExecutable)
	if result.TimedOut {
		return fmt.Errorf("git %s: command timed out after %s", sanitizeGitArgs(result.Args), gitCommandTimeout)
	}
	if result.Err == nil {
		return nil
	}
	return fmt.Errorf("git %s: %s", sanitizeGitArgs(result.Args), sanitizedResultMessage(result))
}

func gitOutput(args ...string) (string, error) {
	result := runGitCaptured(args, gitExecutable)
	if result.TimedOut {
		return "", fmt.Errorf("git %s: command timed out after %s", sanitizeGitArgs(result.Args), gitCommandTimeout)
	}
	if result.Err != nil {
		return "", fmt.Errorf("git %s: %s", sanitizeGitArgs(result.Args), sanitizedResultMessage(result))
	}
	return result.Output, nil
}

func runGitCaptured(args []string, executable string) gitCommandResult {
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()
	output, err := gitCommandRunner(ctx, executable, args, sanitizedGitEnv())
	result := gitCommandResult{Args: append([]string(nil), args...), Output: string(output), Err: err}
	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		return result
	}
	if err == nil {
		return result
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	} else {
		result.ExitCode = -1
	}
	return result
}

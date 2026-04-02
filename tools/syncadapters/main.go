package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	rootFlag := flag.String("root", ".", "repository root path")
	outputFlag := flag.String("output", filepath.Join("build", "generated", "adapters"), "rendered adapter output root")
	toolFlag := flag.String("tool", "", "optional adapter tool id to generate")
	flag.Parse()

	if err := run(*rootFlag, *outputFlag, *toolFlag); err != nil {
		fatalf("sync adapters: %v", err)
	}
}

func fatalf(pattern string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, pattern+"\n", args...)
	os.Exit(1)
}

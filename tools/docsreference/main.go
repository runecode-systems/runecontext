package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/runecode-systems/runecontext/internal/cli"
)

func main() {
	rootFlag := flag.String("root", ".", "repository root path")
	flag.Parse()

	root, err := filepath.Abs(*rootFlag)
	if err != nil {
		fatalf("resolve repository root: %v", err)
	}
	if err := cli.WriteDocumentationReferenceArtifacts(root); err != nil {
		fatalf("write documentation reference artifacts: %v", err)
	}
}

func fatalf(pattern string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, pattern+"\n", args...)
	os.Exit(1)
}

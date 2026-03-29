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
	version, err := cli.ReadReleaseMetadataVersion(root)
	if err != nil {
		fatalf("read release metadata version: %v", err)
	}
	if err := cli.WriteMetadataSyncArtifacts(root, version); err != nil {
		fatalf("write metadata sync artifacts: %v", err)
	}
}

func fatalf(pattern string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, pattern+"\n", args...)
	os.Exit(1)
}

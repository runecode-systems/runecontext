package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestManualRepoInstallFlowOverReferenceFixture(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	bundleRoot := t.TempDir()
	copyDirForCLI(t, filepath.Join(repoRoot, "schemas"), filepath.Join(bundleRoot, "schemas"))
	projectRoot := filepath.Join(bundleRoot, "projects", "embedded")
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), projectRoot)

	originalWD := t.TempDir()
	if cwd, err := filepath.Abs("."); err == nil {
		originalWD = cwd
	}
	t.Cleanup(func() { t.Chdir(originalWD) })
	t.Chdir(bundleRoot)

	var validateOut bytes.Buffer
	var validateErr bytes.Buffer
	if code := Run([]string{"validate", "--path", projectRoot}, &validateOut, &validateErr); code != exitOK {
		t.Fatalf("expected validate success from bundle-style layout, got %d (%s)", code, validateErr.String())
	}
	if !strings.Contains(validateOut.String(), "result=ok") {
		t.Fatalf("expected validate success output, got %q", validateOut.String())
	}

	var doctorOut bytes.Buffer
	var doctorErr bytes.Buffer
	if code := Run([]string{"doctor", "--path", projectRoot}, &doctorOut, &doctorErr); code != exitOK {
		t.Fatalf("expected doctor success from bundle-style layout, got %d (%s)", code, doctorErr.String())
	}
	if !strings.Contains(doctorOut.String(), "command=doctor") {
		t.Fatalf("expected doctor output, got %q", doctorOut.String())
	}
}

func TestQuickInstallConfirmationViaVersionAndDoctor(t *testing.T) {
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v0.1.0-alpha.8"

	var versionOut bytes.Buffer
	var versionErr bytes.Buffer
	if code := Run([]string{"--version"}, &versionOut, &versionErr); code != exitOK {
		t.Fatalf("expected --version success, got %d (%s)", code, versionErr.String())
	}
	fields := parseCLIKeyValueOutput(t, versionOut.String())
	if got, want := fields["version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected normalized version %q, got %q", want, got)
	}

	projectRoot := repoFixtureRoot(t, "reference-projects", "embedded")
	var doctorOut bytes.Buffer
	var doctorErr bytes.Buffer
	if code := Run([]string{"doctor", "--path", projectRoot}, &doctorOut, &doctorErr); code != exitOK {
		t.Fatalf("expected doctor success, got %d (%s)", code, doctorErr.String())
	}
	doctorFields := parseCLIKeyValueOutput(t, doctorOut.String())
	if got, want := doctorFields["result"], "ok"; got != want {
		t.Fatalf("expected doctor result %q, got %q", want, got)
	}
}

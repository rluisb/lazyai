package scaffold

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrintMCPEnvGuidance_PrintsExportsWithoutGeneratingEnvFile(t *testing.T) {
	targetDir := t.TempDir()
	aiDir := filepath.Join(targetDir, ".ai")
	if err := os.MkdirAll(aiDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	mcp := `{
		"servers": {
			"alpha": {"env": {"ALPHA_TOKEN": "${ALPHA_TOKEN}"}},
			"beta": {"env": {"BETA_KEY": "${BETA_KEY}"}},
			"disabled": {"enabled": false, "env": {"DISABLED_KEY": "${DISABLED_KEY}"}}
		}
	}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcp), 0o644); err != nil {
		t.Fatalf("write mcp: %v", err)
	}

	output := captureStdout(t, func() {
		if err := PrintMCPEnvGuidance(targetDir, nil, "", nil); err != nil {
			t.Fatalf("guidance: %v", err)
		}
	})

	for _, want := range []string{
		"LazyAI does not create or manage .env files",
		"export ALPHA_TOKEN=\"\"",
		"export BETA_KEY=\"\"",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output:\n%s", want, output)
		}
	}
	if strings.Contains(output, "DISABLED_KEY") {
		t.Fatalf("disabled server env should not be printed:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".env.example")); !os.IsNotExist(err) {
		t.Fatalf("expected no .env.example generated, stat err: %v", err)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return buf.String()
}

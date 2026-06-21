package cmd

import "testing"

func TestImportWithUnsupportedToolFailsBeforeDetection(t *testing.T) {
	sourceDir := t.TempDir()
	cmd := newImportCommand("gemini", true, true)

	err := runImport(cmd, []string{sourceDir})
	if err == nil || err.Error() != "unsupported tool \"gemini\" (supported tools: antigravity, claude-code, copilot, kiro, omp, opencode, pi)" {
		t.Fatalf("runImport error = %v, want unsupported-tool error", err)
	}
}

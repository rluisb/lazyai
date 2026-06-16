package minimality

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/tokenrent"
)

func TestAnalyzeProjectProducesDeterministicReport(t *testing.T) {
	root := t.TempDir()
	writeMinimalCLICommandSource(t, root)
	writeLargeGoFile(t, root, "pkg/zeta.go", 700)
	writeLargeGoFile(t, root, "pkg/alpha.go", 700)
	writeLargeGoFile(t, root, "pkg/middle.go", 501)
	writeLargeGoFile(t, root, "pkg/ignored.go", 500)
	writeCanonicalAsset(t, root, "agents/implementer.md", "implementer")
	writeCanonicalAsset(t, root, "skills/diagnose.md", "diagnose")
	writeCanonicalAsset(t, root, "hooks/pre-commit.md", "hook")
	writeCanonicalAsset(t, root, "templates/bug.md", "template")
	writeCanonicalAsset(t, root, "rules/go.md", "rule")

	first, err := AnalyzeProject(root)
	if err != nil {
		t.Fatalf("AnalyzeProject returned error: %v", err)
	}
	second, err := AnalyzeProject(root)
	if err != nil {
		t.Fatalf("second AnalyzeProject returned error: %v", err)
	}

	if got := first.GoFiles; len(got) != 3 {
		t.Fatalf("GoFiles length = %d, want 3: %#v", len(got), got)
	}
	wantOrder := []string{"pkg/alpha.go", "pkg/zeta.go", "pkg/middle.go"}
	for i, want := range wantOrder {
		if first.GoFiles[i].Path != want {
			t.Fatalf("GoFiles[%d].Path = %q, want %q", i, first.GoFiles[i].Path, want)
		}
	}
	if first.CLICommands.Visible != 1 || first.CLICommands.Hidden != 1 || first.CLICommands.Registered != 2 {
		t.Fatalf("CLICommands = %#v", first.CLICommands)
	}
	if first.CanonicalLibrary.Bytes != len("implementer")+len("diagnose")+len("hook")+len("template")+len("rule") {
		t.Fatalf("CanonicalLibrary.Bytes = %d", first.CanonicalLibrary.Bytes)
	}
	if first.CanonicalLibrary.BudgetBytes != tokenrent.DefaultBudgetBytes {
		t.Fatalf("CanonicalLibrary.BudgetBytes = %d", first.CanonicalLibrary.BudgetBytes)
	}
	assertAssetCount(t, first, "agents", 1, true)
	assertAssetCount(t, first, "skills", 1, true)
	assertAssetCount(t, first, "hooks", 1, true)
	assertAssetCount(t, first, "prompts", 0, false)
	assertAssetCount(t, first, "templates", 1, true)
	assertAssetCount(t, first, "rules", 1, true)

	var firstOut bytes.Buffer
	var secondOut bytes.Buffer
	if err := WriteText(&firstOut, first); err != nil {
		t.Fatalf("WriteText first: %v", err)
	}
	if err := WriteText(&secondOut, second); err != nil {
		t.Fatalf("WriteText second: %v", err)
	}
	if firstOut.String() != secondOut.String() {
		t.Fatalf("report output is not deterministic\nfirst:\n%s\nsecond:\n%s", firstOut.String(), secondOut.String())
	}
	if !strings.Contains(firstOut.String(), "Mode: report-only") {
		t.Fatalf("report should state report-only mode:\n%s", firstOut.String())
	}
}

func TestRunIsReportOnlyWhenThresholdsAreExceeded(t *testing.T) {
	root := t.TempDir()
	writeMinimalCLICommandSource(t, root)
	writeLargeGoFile(t, root, "pkg/large.go", 501)
	writeCanonicalAsset(t, root, "agents/large.md", string(make([]byte, tokenrent.DefaultBudgetBytes+1)))

	var out bytes.Buffer
	if err := Run(root, &out); err != nil {
		t.Fatalf("Run returned error for report-only threshold breach: %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "pkg/large.go") {
		t.Fatalf("report missing large Go file:\n%s", text)
	}
	if !strings.Contains(text, "token-rent budget: 50000") {
		t.Fatalf("report missing token-rent budget:\n%s", text)
	}
	if !strings.Contains(text, "Canonical bytes are informational here; token-rent remains the hard gate.") {
		t.Fatalf("report missing token-rent hard gate note:\n%s", text)
	}
}

func writeMinimalCLICommandSource(t *testing.T, root string) {
	t.Helper()
	content := `package cmd

import "github.com/spf13/cobra"

var visibleCmd = &cobra.Command{Use: "visible"}
var hiddenCmd = &cobra.Command{Use: "hidden", Hidden: true}

func init() {
	rootCmd.AddCommand(hiddenCmd)
	rootCmd.AddCommand(visibleCmd)
}
`
	writeFile(t, root, filepath.Join("packages", "cli", "cmd", "commands.go"), content)
}

func writeLargeGoFile(t *testing.T, root, rel string, lines int) {
	t.Helper()
	var builder strings.Builder
	builder.WriteString("package sample\n")
	for i := 1; i < lines; i++ {
		builder.WriteString("// filler\n")
	}
	writeFile(t, root, rel, builder.String())
}

func writeCanonicalAsset(t *testing.T, root, rel, content string) {
	t.Helper()
	writeFile(t, root, filepath.Join(tokenrent.CanonicalSubdir, rel), content)
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func assertAssetCount(t *testing.T, report Report, category string, count int, present bool) {
	t.Helper()
	for _, asset := range report.CanonicalAssets {
		if asset.Category != category {
			continue
		}
		if asset.Count != count || asset.Present != present {
			t.Fatalf("asset %s = %#v, want count=%d present=%v", category, asset, count, present)
		}
		return
	}
	t.Fatalf("asset category %s not found in %#v", category, report.CanonicalAssets)
}

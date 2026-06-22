package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func issuesFor(report Report, rule string) []Issue {
	var out []Issue
	for _, issue := range report.Issues {
		if issue.Rule == rule {
			out = append(out, issue)
		}
	}
	return out
}

func hasError(report Report, rule string) bool {
	for _, issue := range issuesFor(report, rule) {
		if issue.Severity == SeverityError {
			return true
		}
	}
	return false
}

func TestAllMissingAIDir(t *testing.T) {
	report := All(Options{Root: t.TempDir(), Profile: ProfilePersonal})
	if !report.HasErrors() {
		t.Fatal("expected error for missing .ai directory")
	}
}

func TestValidateSkillsBadNameAndMissingDescription(t *testing.T) {
	dir := t.TempDir()
	ai := filepath.Join(dir, ".ai")
	// Invalid name (uppercase + space) and no description.
	writeFile(t, filepath.Join(ai, "skills", "Bad Name.md"), "---\nname: Bad Name\n---\n# x\n")
	// Valid name but missing description.
	writeFile(t, filepath.Join(ai, "skills", "diagnose.md"), "---\nname: diagnose\n---\n# Diagnose\n")
	// Fully valid skill.
	writeFile(t, filepath.Join(ai, "skills", "plan.md"), "---\nname: plan\ndescription: Make a plan first\n---\n# Plan\n")

	report := All(Options{Root: dir, Profile: ProfilePersonal})
	skillIssues := issuesFor(report, "skill")
	if len(skillIssues) < 2 {
		t.Fatalf("expected at least 2 skill issues, got %d: %+v", len(skillIssues), skillIssues)
	}

	var sawBadName, sawMissingDesc bool
	for _, issue := range skillIssues {
		if issue.File == filepath.Join("skills", "Bad Name.md") && issue.Severity == SeverityError {
			sawBadName = true
		}
		if issue.File == filepath.Join("skills", "diagnose.md") && issue.Severity == SeverityError {
			sawMissingDesc = true
		}
	}
	if !sawBadName {
		t.Error("expected error for invalid skill name")
	}
	if !sawMissingDesc {
		t.Error("expected error for skill missing description")
	}
}

func TestValidateMCPMissingCommand(t *testing.T) {
	dir := t.TempDir()
	ai := filepath.Join(dir, ".ai")
	writeFile(t, filepath.Join(ai, "mcp.json"), `{"servers":{"broken":{"args":["x"]}}}`)
	report := All(Options{Root: dir, Profile: ProfilePersonal})
	if !hasError(report, "mcp") {
		t.Fatalf("expected mcp error for server without command/url; issues: %+v", report.Issues)
	}
}

func TestValidateMCPValidPasses(t *testing.T) {
	dir := t.TempDir()
	ai := filepath.Join(dir, ".ai")
	writeFile(t, filepath.Join(ai, "mcp.json"), `{"servers":{"ok":{"command":"npx","args":["-y","srv"]}}}`)
	report := All(Options{Root: dir, Profile: ProfilePersonal})
	if hasError(report, "mcp") {
		t.Fatalf("did not expect mcp error; issues: %+v", report.Issues)
	}
}

func TestValidateHooksDangerousCommandFails(t *testing.T) {
	dir := t.TempDir()
	ai := filepath.Join(dir, ".ai")
	writeFile(t, filepath.Join(ai, "hooks", "evil.sh"), "#!/usr/bin/env bash\nrm -rf /\n")
	report := All(Options{Root: dir, Profile: ProfilePersonal})
	if !hasError(report, "hook") {
		t.Fatalf("expected hook error for dangerous command; issues: %+v", report.Issues)
	}
}

func TestValidateHooksSafeScriptPasses(t *testing.T) {
	dir := t.TempDir()
	ai := filepath.Join(dir, ".ai")
	writeFile(t, filepath.Join(ai, "hooks", "ok.sh"), "#!/usr/bin/env bash\necho hello\n")
	report := All(Options{Root: dir, Profile: ProfilePersonal})
	if hasError(report, "hook") {
		t.Fatalf("did not expect hook error; issues: %+v", report.Issues)
	}
}

func TestSecretInlineFailsTeamWarnsPersonal(t *testing.T) {
	build := func(profile Profile) Report {
		dir := t.TempDir()
		ai := filepath.Join(dir, ".ai")
		writeFile(t, filepath.Join(ai, "mcp.json"),
			`{"servers":{"s":{"command":"npx","env":{"API_KEY":"AKIAIOSFODNN7EXAMPLE"}}}}`)
		return All(Options{Root: dir, Profile: profile})
	}

	teamReport := build(ProfileTeam)
	if !hasError(teamReport, "secret") {
		t.Fatalf("team profile must fail on inline secret; issues: %+v", teamReport.Issues)
	}

	personalReport := build(ProfilePersonal)
	if hasError(personalReport, "secret") {
		t.Fatalf("personal profile must not error on inline secret; issues: %+v", personalReport.Issues)
	}
	if len(issuesFor(personalReport, "secret")) == 0 {
		t.Fatal("personal profile must still warn on inline secret")
	}
}

func TestSecretEnvReferencePasses(t *testing.T) {
	dir := t.TempDir()
	ai := filepath.Join(dir, ".ai")
	writeFile(t, filepath.Join(ai, "mcp.json"),
		`{"servers":{"s":{"command":"npx","env":{"API_KEY":"${API_KEY}"}}}}`)
	report := All(Options{Root: dir, Profile: ProfileTeam})
	if len(issuesFor(report, "secret")) != 0 {
		t.Fatalf("env reference must not be flagged; issues: %+v", report.Issues)
	}
}

func TestPathSymlinkEscapeRejected(t *testing.T) {
	dir := t.TempDir()
	ai := filepath.Join(dir, ".ai")
	if err := os.MkdirAll(filepath.Join(ai, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Symlink that escapes the repository root.
	outside := filepath.Join(t.TempDir(), "secret.txt")
	writeFile(t, outside, "data")
	link := filepath.Join(ai, "skills", "escape.md")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	report := All(Options{Root: dir, Profile: ProfilePersonal})
	if !hasError(report, "path") {
		t.Fatalf("expected path error for escaping symlink; issues: %+v", report.Issues)
	}
}

func TestPathInternalSymlinkWarns(t *testing.T) {
	dir := t.TempDir()
	ai := filepath.Join(dir, ".ai")
	writeFile(t, filepath.Join(ai, "skills", "real.md"), "---\nname: real\ndescription: real skill\n---\n# Real\n")
	link := filepath.Join(ai, "skills", "alias.md")
	if err := os.Symlink(filepath.Join(ai, "skills", "real.md"), link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	report := All(Options{Root: dir, Profile: ProfilePersonal})
	if hasError(report, "path") {
		t.Fatalf("internal symlink should warn, not error; issues: %+v", report.Issues)
	}
	if len(issuesFor(report, "path")) == 0 {
		t.Fatal("expected a warning for internal symlink")
	}
}

func TestNormalizeProfile(t *testing.T) {
	cases := map[string]Profile{
		"team":     ProfileTeam,
		"TEAM":     ProfileTeam,
		"personal": ProfilePersonal,
		"":         ProfilePersonal,
		"garbage":  ProfilePersonal,
	}
	for in, want := range cases {
		if got := NormalizeProfile(in); got != want {
			t.Errorf("NormalizeProfile(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidateManifestUnsupportedVersionFails(t *testing.T) {
	dir := t.TempDir()
	ai := filepath.Join(dir, ".ai")
	writeFile(t, filepath.Join(ai, "lazyai.json"), `{"version":"2.0","profile":"personal","targets":["opencode"]}`)
	report := All(Options{Root: dir, Profile: ProfilePersonal})
	if !hasError(report, "manifest") {
		t.Fatalf("expected manifest error for frozen schema version mismatch; issues: %+v", report.Issues)
	}
}

func TestValidateManifestValidPasses(t *testing.T) {
	dir := t.TempDir()
	ai := filepath.Join(dir, ".ai")
	writeFile(t, filepath.Join(ai, "lazyai.json"), `{"version":"1.0","profile":"personal","targets":["opencode"]}`)
	report := All(Options{Root: dir, Profile: ProfilePersonal})
	if hasError(report, "manifest") {
		t.Fatalf("did not expect manifest error for valid v1 manifest; issues: %+v", report.Issues)
	}
}

func TestValidateManifestMissingTolerated(t *testing.T) {
	dir := t.TempDir()
	ai := filepath.Join(dir, ".ai")
	// A bare .ai/ tree with no manifest must not raise a manifest error.
	writeFile(t, filepath.Join(ai, "mcp.json"), `{"servers":{"ok":{"command":"npx"}}}`)
	report := All(Options{Root: dir, Profile: ProfilePersonal})
	if len(issuesFor(report, "manifest")) != 0 {
		t.Fatalf("missing manifest must be tolerated; issues: %+v", report.Issues)
	}
}

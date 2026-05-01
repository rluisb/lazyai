package validation

import (
	"testing"
)

func TestValidateSpec006Metadata(t *testing.T) {
	t.Parallel()

	valid := map[string]any{
		"schema_version":        1,
		"artifact_type":         "spec_plan",
		"id":                    "spec-006-plan",
		"title":                 "Spec 006 Plan",
		"created_at":            "2026-04-17T10:00:00Z",
		"updated_at":            "2026-04-17T10:05:00Z",
		"created_by":            "planner",
		"updated_by":            "planner",
		"workflow_id":           "rpi",
		"workflow_run_id":       "run-1",
		"related_document_refs": []any{"specs/006/plan.md"},
	}

	if issues := ValidateSpec006Metadata(valid); len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}

	invalid := map[string]any{
		"schema_version":        1,
		"created_at":            "bad-date",
		"updated_at":            "2026-04-16T10:00:00Z",
		"created_by":            "planner",
		"updated_by":            "planner",
		"workflow_run_id":       "run-1",
		"related_document_refs": []any{"/absolute/path.md"},
	}

	issues := ValidateSpec006Metadata(invalid)
	if len(issues) < 4 {
		t.Fatalf("expected multiple issues, got %v", issues)
	}
}

func TestValidateArtifactName_AcceptsValidNames(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"my-agent",
		"builder",
		"code-review",
		"test123",
		"abc",
	}

	for _, name := range validNames {
		if err := ValidateArtifactName(name); err != nil {
			t.Errorf("ValidateArtifactName(%q) returned error: %v", name, err)
		}
	}
}

func TestValidateArtifactName_RejectsInvalidNames(t *testing.T) {
	t.Parallel()

	invalidCases := []struct {
		name string
		desc string
	}{
		{"", "empty"},
		{"My-Agent", "uppercase"},
		{"my agent", "spaces"},
		{"my_agent", "underscores"},
		{"1start-with-number", "starts with number"},
		{"a", "too short"},
	}

	for _, tc := range invalidCases {
		err := ValidateArtifactName(tc.name)
		if err == nil {
			t.Errorf("ValidateArtifactName(%q) should reject (%s)", tc.name, tc.desc)
		}
	}
}

func TestSanitizeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"My Agent", "my-agent"},
		{"code_review", "code-review"},
		{"UPPER CASE", "upper-case"},
		{"  spaces  ", "spaces"},
		{"special<>:chars", "specialchars"},
		{"multiple---dashes", "multiple-dashes"},
		{"My Cool Agent", "my-cool-agent"},
	}

	for _, tt := range tests {
		got := SanitizeName(tt.input)
		if got != tt.want {
			t.Errorf("SanitizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestValidateToolId(t *testing.T) {
	t.Parallel()

	validTools := []string{"opencode", "claude-code", "copilot"}
	for _, tool := range validTools {
		if err := ValidateToolId(tool); err != nil {
			t.Errorf("ValidateToolId(%q) returned error: %v", tool, err)
		}
	}

	if err := ValidateToolId("invalid-tool"); err == nil {
		t.Error("ValidateToolId should reject unknown tool")
	}
}

func TestIsValidArtifactType(t *testing.T) {
	t.Parallel()

	validTypes := []string{"agent", "skill", "command", "prompt", "template", "workflow", "domain", "mode"}
	for _, typ := range validTypes {
		if !IsValidArtifactType(typ) {
			t.Errorf("IsValidArtifactType(%q) = false, want true", typ)
		}
	}

	if IsValidArtifactType("invalid") {
		t.Error("IsValidArtifactType(\"invalid\") = true, want false")
	}
}

func TestValidateRequiredText(t *testing.T) {
	t.Parallel()

	if err := ValidateRequiredText("hello", "field"); err != nil {
		t.Errorf("ValidateRequiredText(hello) returned error: %v", err)
	}
	if err := ValidateRequiredText("", "field"); err == nil {
		t.Error("ValidateRequiredText(\"\") should return error")
	}
	if err := ValidateRequiredText("   ", "field"); err == nil {
		t.Error("ValidateRequiredText(\"   \") should return error")
	}
}

func TestValidateFilesystemSafeName(t *testing.T) {
	t.Parallel()

	if err := ValidateFilesystemSafeName("valid-name", "test"); err != nil {
		t.Errorf("ValidateFilesystemSafeName(valid-name) returned error: %v", err)
	}
	if err := ValidateFilesystemSafeName("", "test"); err == nil {
		t.Error("empty name should be rejected")
	}
	if err := ValidateFilesystemSafeName("bad<>name", "test"); err == nil {
		t.Error("name with invalid chars should be rejected")
	}
}

package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestFeatureChainSourceShapeIsSequentialAndGated(t *testing.T) {
	data := readSourceFeatureChain(t)
	assertCurrentFeatureChainShape(t, data)
}

func TestScaffoldOrchestration_InstallsFeatureChainShapeFromLibrary(t *testing.T) {
	targetDir := t.TempDir()
	libFS := createMinimalLibraryFS()
	data := readSourceFeatureChain(t)
	libFS["orchestration/chains/feature.json"] = &fstest.MapFile{Data: data}
	fileRecords := []types.TrackedFile{}

	if err := ScaffoldOrchestration(targetDir, libFS, &fileRecords, types.ConflictStrategySkip, map[string]types.ConflictStrategy{}); err != nil {
		t.Fatalf("ScaffoldOrchestration failed: %v", err)
	}

	installed, err := os.ReadFile(filepath.Join(targetDir, ".ai", "orchestration", "chains", "feature.json"))
	if err != nil {
		t.Fatalf("read installed feature chain: %v", err)
	}
	assertCurrentFeatureChainShape(t, installed)
	if string(installed) != string(data) {
		t.Fatal("expected scaffolded feature chain to match the source chain byte-for-byte")
	}
}

func TestScaffoldOrchestration_CopiesChainDefinitionsWithoutTemplateRendering(t *testing.T) {
	targetDir := t.TempDir()
	libFS := createMinimalLibraryFS()
	chainWithTemplateDirectives := []byte(`{
  "kind": "chain",
  "name": "feature",
  "entry": "plan-quality",
  "steps": [
    {"id":"plan-quality","agent":"planner","skills":["plan"],"description":"quality","transitions":{"success":"red-team-plan"}},
    {{#if features.adversarialDesign}}
    {"id":"red-team-plan","agent":"reviewer","skills":["red-team-plan"],"description":"red team","transitions":{"success":"plan-gate"}},
    {{/if}}
    {"id":"plan-gate","agent":"planner","skills":[],"description":"gate","gate":"user_approval","transitions":{"approved":"implement","rejected":"plan-quality"}}
  ]
}`)
	libFS["orchestration/chains/feature.json"] = &fstest.MapFile{Data: chainWithTemplateDirectives}
	fileRecords := []types.TrackedFile{}

	if err := ScaffoldOrchestration(targetDir, libFS, &fileRecords, types.ConflictStrategySkip, map[string]types.ConflictStrategy{}); err != nil {
		t.Fatalf("ScaffoldOrchestration failed: %v", err)
	}

	installed, err := os.ReadFile(filepath.Join(targetDir, ".ai", "orchestration", "chains", "feature.json"))
	if err != nil {
		t.Fatalf("read installed feature chain: %v", err)
	}

	if string(installed) != string(chainWithTemplateDirectives) {
		t.Fatal("expected orchestration scaffold to copy chain definitions without rendering feature templates")
	}
	if !strings.Contains(string(installed), "{{#if features.adversarialDesign}}") {
		t.Fatal("expected unsupported template directive to remain literal in installed chain")
	}
}

func TestScaffoldOrchestration_AddsExtensionOrchestrationContent(t *testing.T) {
	targetDir := t.TempDir()
	libFS := createMinimalLibraryFS()
	fileRecords := []types.TrackedFile{}

	localExtDir := filepath.Join(targetDir, ".ai", "extensions", "team-pack")
	if err := os.MkdirAll(filepath.Join(localExtDir, "skills"), 0o755); err != nil {
		t.Fatalf("mkdir local extension skills: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(localExtDir, "orchestration", "chains"), 0o755); err != nil {
		t.Fatalf("mkdir local extension chains: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localExtDir, "skills", "custom.md"), []byte("# custom"), 0o644); err != nil {
		t.Fatalf("write local extension skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localExtDir, "orchestration", "chains", "release.json"), []byte(`{"name":"release"}`), 0o644); err != nil {
		t.Fatalf("write local extension chain: %v", err)
	}

	sharedExtDir := filepath.Join(targetDir, "shared-ext")
	if err := os.MkdirAll(filepath.Join(sharedExtDir, "skills"), 0o755); err != nil {
		t.Fatalf("mkdir shared extension skills: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(sharedExtDir, "orchestration", "workflows"), 0o755); err != nil {
		t.Fatalf("mkdir shared extension workflows: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sharedExtDir, "skills", "deploy.md"), []byte("# deploy"), 0o644); err != nil {
		t.Fatalf("write shared extension skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sharedExtDir, "orchestration", "workflows", "deploy.json"), []byte(`{"name":"deploy"}`), 0o644); err != nil {
		t.Fatalf("write shared extension workflow: %v", err)
	}

	config := "[extensions.shared]\npath = \"./shared-ext\"\n"
	if err := os.WriteFile(filepath.Join(targetDir, ".ai-setup.toml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write extension config: %v", err)
	}

	if err := ScaffoldOrchestration(targetDir, libFS, &fileRecords, types.ConflictStrategySkip, map[string]types.ConflictStrategy{}); err != nil {
		t.Fatalf("ScaffoldOrchestration failed: %v", err)
	}

	if !filesExist(
		filepath.Join(targetDir, ".ai", "orchestration", "chains", "feature.json"),
		filepath.Join(targetDir, ".ai", "orchestration", "chains", "release.json"),
		filepath.Join(targetDir, ".ai", "orchestration", "workflows", "deploy.json"),
	) {
		t.Fatal("expected built-in and extension orchestration files to be scaffolded")
	}

	if !hasTrackedFile(fileRecords, ".ai/orchestration/chains/release.json") {
		t.Fatal("expected extension chain to be tracked")
	}
	if !hasTrackedFile(fileRecords, ".ai/orchestration/workflows/deploy.json") {
		t.Fatal("expected extension workflow to be tracked")
	}
}

func readSourceFeatureChain(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "library", "orchestration", "chains", "feature.json"))
	if err != nil {
		t.Fatalf("read source feature chain: %v", err)
	}
	return data
}

func assertCurrentFeatureChainShape(t *testing.T, data []byte) {
	t.Helper()

	raw := string(data)
	for _, forbidden := range []string{"red-team-plan", "{{#if", "{{/if}}"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("feature chain must not contain %q in the D6-only base chain", forbidden)
		}
	}

	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("feature chain must be valid JSON: %v", err)
	}
	if body["kind"] != "chain" {
		t.Fatalf("expected feature chain kind to be chain, got %v", body["kind"])
	}
	if body["name"] != "feature" {
		t.Fatalf("expected feature chain name to be feature, got %v", body["name"])
	}
	if _, ok := body["parallel"]; ok {
		t.Fatal("feature chain must remain sequential and must not define a parallel block")
	}

	steps, ok := body["steps"].([]any)
	if !ok || len(steps) == 0 {
		t.Fatal("feature chain must define a non-empty sequential steps array")
	}

	ids := make([]string, 0, len(steps))
	stepByID := map[string]map[string]any{}
	var plan map[string]any
	var approvalGatesBeforeImplement []string
	for _, rawStep := range steps {
		step, ok := rawStep.(map[string]any)
		if !ok {
			t.Fatalf("feature chain step must be an object, got %T", rawStep)
		}
		for _, forbiddenKey := range []string{"condition", "optionalByFeature", "parallel"} {
			if _, ok := step[forbiddenKey]; ok {
				t.Fatalf("feature chain step %v must not define unsupported %q metadata", step["id"], forbiddenKey)
			}
		}
		id, ok := step["id"].(string)
		if !ok || id == "" {
			t.Fatalf("feature chain step must have a string id, got %v", step["id"])
		}
		ids = append(ids, id)
		stepByID[id] = step
		if id != "implement" && step["gate"] == "user_approval" {
			approvalGatesBeforeImplement = append(approvalGatesBeforeImplement, id)
		}
		if id == "implement" {
			approvalGatesBeforeImplement = append(approvalGatesBeforeImplement, "__implement_boundary__")
		}
		if id == "plan" {
			plan = step
		}
	}

	expectedIDs := []string{"research", "plan", "plan-quality", "plan-gate", "implement", "review", "fix", "document"}
	if strings.Join(ids, ",") != strings.Join(expectedIDs, ",") {
		t.Fatalf("expected current feature chain step order %v, got %v", expectedIDs, ids)
	}
	if plan == nil {
		t.Fatal("expected current feature chain to include a plan step")
	}
	if _, ok := plan["gate"]; ok {
		t.Fatalf("expected plan step not to own user_approval gate, got %v", plan["gate"])
	}

	transitions, ok := plan["transitions"].(map[string]any)
	if !ok {
		t.Fatalf("expected plan transitions object, got %T", plan["transitions"])
	}
	if transitions["success"] != "plan-quality" {
		t.Fatalf("expected plan success transition to plan-quality, got %v", transitions)
	}

	planQuality := stepByID["plan-quality"]
	if planQuality == nil {
		t.Fatal("expected current feature chain to include a plan-quality step")
	}
	qualityTransitions, ok := planQuality["transitions"].(map[string]any)
	if !ok {
		t.Fatalf("expected plan-quality transitions object, got %T", planQuality["transitions"])
	}
	for _, outcome := range []string{"success", "pass", "warn", "fail"} {
		if qualityTransitions[outcome] != "plan-gate" {
			t.Fatalf("expected plan-quality %s transition to plan-gate with no automatic loop, got %v", outcome, qualityTransitions)
		}
	}

	planGate := stepByID["plan-gate"]
	if planGate == nil {
		t.Fatal("expected current feature chain to include a plan-gate step")
	}
	if planGate["gate"] != "user_approval" {
		t.Fatalf("expected plan-gate step to own user_approval gate, got %v", planGate["gate"])
	}
	gateTransitions, ok := planGate["transitions"].(map[string]any)
	if !ok {
		t.Fatalf("expected plan-gate transitions object, got %T", planGate["transitions"])
	}
	if gateTransitions["approved"] != "implement" || gateTransitions["rejected"] != "plan" {
		t.Fatalf("expected explicit plan-gate transitions approved->implement and rejected->plan, got %v", gateTransitions)
	}
	if strings.Join(approvalGatesBeforeImplement, ",") != "plan-gate,__implement_boundary__" {
		t.Fatalf("expected plan-gate to be the only approval gate before implementation, got %v", approvalGatesBeforeImplement)
	}
}

func filesExist(paths ...string) bool {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}

func hasTrackedFile(records []types.TrackedFile, path string) bool {
	for _, record := range records {
		if record.Path == path {
			return true
		}
	}
	return false
}

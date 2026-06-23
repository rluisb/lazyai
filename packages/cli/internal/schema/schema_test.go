package schema

import (
	"encoding/json"
	"testing"
)

func TestAccessorsReturnJSON(t *testing.T) {
	t.Run("lazyai", func(t *testing.T) {
		data := LazyAI()
		if len(data) == 0 {
			t.Fatal("lazyai schema payload is empty")
		}
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("lazyai schema not valid JSON: %v", err)
		}
		required, ok := payload["required"].([]any)
		if !ok {
			t.Fatal("lazyai schema missing required array")
		}
		if !containsString(required, "version") || !containsString(required, "targets") {
			t.Fatalf("lazyai schema required must include version and targets, got %#v", required)
		}
		targetsEnum := mapPathSlice(payload, []string{"properties", "targets", "items", "enum"})
		if len(targetsEnum) != 8 {
			t.Fatalf("expected 8 manifest targets, got %d", len(targetsEnum))
		}
		reqTargets := []string{"opencode", "claude", "claude-code", "copilot", "pi", "omp", "antigravity", "kiro"}
		for _, want := range reqTargets {
			if !containsString(targetsEnum, want) {
				t.Fatalf("manifest enum missing target %q", want)
			}
		}
		if containsString(targetsEnum, "codex") {
			t.Fatal("manifest targets enum must not contain codex")
		}
	})

	t.Run("lock", func(t *testing.T) {
		data := Lock()
		if len(data) == 0 {
			t.Fatal("lock schema payload is empty")
		}
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("lock schema not valid JSON: %v", err)
		}
	})

	t.Run("mcp", func(t *testing.T) {
		data := MCPCatalog()
		if len(data) == 0 {
			t.Fatal("mcp schema payload is empty")
		}
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("mcp schema not valid JSON: %v", err)
		}
	})

	t.Run("eval case", func(t *testing.T) {
		data := EvalCase()
		if len(data) == 0 {
			t.Fatal("eval-case schema payload is empty")
		}
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("eval-case schema not valid JSON: %v", err)
		}
		required, ok := payload["required"].([]any)
		if !ok {
			t.Fatal("eval-case schema missing required array")
		}
		for _, field := range []string{"id", "title", "input", "expected"} {
			if !containsString(required, field) {
				t.Fatalf("eval-case required must include %q", field)
			}
		}
	})

	t.Run("eval holdout", func(t *testing.T) {
		data := EvalHoldout()
		if len(data) == 0 {
			t.Fatal("eval-holdout schema payload is empty")
		}
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("eval-holdout schema not valid JSON: %v", err)
		}
		required, ok := payload["required"].([]any)
		if !ok {
			t.Fatal("eval-holdout schema missing required array")
		}
		for _, field := range []string{"id", "title", "input", "expected"} {
			if !containsString(required, field) {
				t.Fatalf("eval-holdout required must include %q", field)
			}
		}
	})

	t.Run("eval rubric", func(t *testing.T) {
		data := EvalRubric()
		if len(data) == 0 {
			t.Fatal("eval-rubric schema payload is empty")
		}
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("eval-rubric schema not valid JSON: %v", err)
		}
		if payload["type"] != "object" {
			t.Fatalf("expected eval-rubric schema type object, got %v", payload["type"])
		}
		required, ok := payload["required"].([]any)
		if !ok {
			t.Fatal("eval-rubric schema missing required array")
		}
		for _, field := range []string{"id", "title", "criteria"} {
			if !containsString(required, field) {
				t.Fatalf("eval-rubric required must include %q", field)
			}
		}
		props, ok := payload["properties"].(map[string]any)
		if !ok {
			t.Fatal("eval-rubric schema missing properties")
		}
		criteria, ok := props["criteria"].(map[string]any)
		if !ok {
			t.Fatal("eval-rubric schema missing criteria property")
		}
		if criteria["type"] != "array" {
			t.Fatalf("expected criteria type array, got %v", criteria["type"])
		}
		items, ok := criteria["items"].(map[string]any)
		if !ok {
			t.Fatal("eval-rubric criteria missing items")
		}
		criteriaRequired, ok := items["required"].([]any)
		if !ok {
			t.Fatal("eval-rubric criteria items missing required array")
		}
		for _, field := range []string{"id", "label", "weight", "description", "pass", "fail"} {
			if !containsString(criteriaRequired, field) {
				t.Fatalf("criteria items required must include %q", field)
			}
		}
	})

	t.Run("names", func(t *testing.T) {
		names := Names()
		if len(names) != 6 {
			t.Fatalf("expected 6 schema names, got %d", len(names))
		}
	})
}

func mapPathSlice(root map[string]any, path []string) []any {
	var cur any = root
	for _, key := range path {
		node, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = node[key]
	}
	arr, _ := cur.([]any)
	return arr
}

func containsString(list []any, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

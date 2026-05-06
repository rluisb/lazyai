package dashboard

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestContractJSONShapes(t *testing.T) {
	overview := DashboardOverview{
		Health: HealthView{
			Status:        "ok",
			Name:          "lazyai-orchestrator",
			Port:          8765,
			PID:           123,
			StartedAt:     "2026-05-05T10:00:00Z",
			ProjectRoot:   "/repo",
			Scope:         "project",
			ExecutionMode: "native",
			ActiveRuns:    db.ActiveRunCounts{Chains: 1, Total: 1},
		},
		ActiveRuns:       db.ActiveRunCounts{Chains: 1, Total: 1},
		RunCountsByState: map[string]int{"running": 1},
		RecentRuns: []RunSummary{{
			Kind:              types.RunKindChain,
			ID:                "chain-1",
			DefinitionName:    "release",
			DefinitionVersion: "2",
			State:             "running",
			Current:           "implement",
			ProjectRoot:       "/repo",
			CreatedAt:         "2026-05-05T10:00:00Z",
			UpdatedAt:         "2026-05-05T10:01:00Z",
			BudgetHealth:      string(types.HealthWarning),
			ErrorCount:        1,
		}},
		RecentErrors: []ErrorEntry{{
			ID:             "err-1",
			RunID:          "chain-1",
			RunKind:        types.RunKindChain,
			DefinitionName: "release",
			StepID:         "implement",
			Category:       string(types.ErrorTransient),
			Code:           "dispatch_failed",
			Message:        "dispatch failed",
			CreatedAt:      "2026-05-05T10:02:00Z",
		}},
		CatalogCounts: CatalogCounts{Total: 2, ByKind: map[string]int{"chain": 1, "team": 1}},
		GeneratedAt:   "2026-05-05T10:03:00Z",
	}

	encoded, err := json.Marshal(overview)
	if err != nil {
		t.Fatalf("marshal overview: %v", err)
	}
	jsonText := string(encoded)
	for _, field := range []string{"\"health\"", "\"activeRuns\"", "\"runCountsByState\"", "\"recentRuns\"", "\"recentErrors\"", "\"catalogCounts\"", "\"generatedAt\""} {
		if !strings.Contains(jsonText, field) {
			t.Fatalf("overview JSON missing %s: %s", field, jsonText)
		}
	}

	detail := RunDetail{
		Summary: overview.RecentRuns[0],
		State:   map[string]any{"state": "running"},
		Steps:   []types.StepState{{StepID: "implement", State: types.StepRunning}},
		Budget: &BudgetView{
			State:      &types.BudgetState{PolicyID: "default", Scope: "chain", ByStep: map[string]types.StepUsage{"implement": {TotalTokens: 10}}},
			Evaluation: &types.BudgetEvaluation{Overall: types.HealthOK},
			ByStep:     map[string]types.StepUsage{"implement": {TotalTokens: 10}},
		},
		Events: []DashboardEvent{{ID: 1, RunID: "chain-1", EventType: "started", Data: map[string]any{"stepId": "implement"}, CreatedAt: "2026-05-05T10:00:00Z"}},
		Errors: overview.RecentErrors,
	}
	encoded, err = json.Marshal(detail)
	if err != nil {
		t.Fatalf("marshal detail: %v", err)
	}
	jsonText = string(encoded)
	for _, field := range []string{"\"summary\"", "\"state\"", "\"steps\"", "\"budget\"", "\"events\"", "\"errors\""} {
		if !strings.Contains(jsonText, field) {
			t.Fatalf("detail JSON missing %s: %s", field, jsonText)
		}
	}
}

func TestContractErrorAndLimitShape(t *testing.T) {
	payload := DashboardErrorResponse{Error: DashboardError{Code: "not_found", Message: "run missing"}}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if got, want := string(encoded), `{"error":{"code":"not_found","message":"run missing"}}`; got != want {
		t.Fatalf("error JSON mismatch\nwant %s\n got %s", want, got)
	}

	if got := NormalizeLimit(0, DefaultRunLimit, MaxRunLimit); got != 50 {
		t.Fatalf("default run limit = %d, want 50", got)
	}
	if got := NormalizeLimit(500, DefaultRunLimit, MaxRunLimit); got != 200 {
		t.Fatalf("max run limit = %d, want 200", got)
	}
	if got := NormalizeLimit(-1, DefaultErrorLimit, MaxErrorLimit); got != 25 {
		t.Fatalf("default error limit = %d, want 25", got)
	}
	if got := NormalizeLimit(500, DefaultErrorLimit, MaxErrorLimit); got != 100 {
		t.Fatalf("max error limit = %d, want 100", got)
	}
}

func TestContractsDoNotExposeWriteOrAdminFields(t *testing.T) {
	contractTypes := []reflect.Type{
		reflect.TypeOf(DashboardOverview{}),
		reflect.TypeOf(RunSummary{}),
		reflect.TypeOf(RunDetail{}),
		reflect.TypeOf(DashboardEvent{}),
		reflect.TypeOf(BudgetView{}),
		reflect.TypeOf(CatalogSummary{}),
		reflect.TypeOf(CatalogDetail{}),
		reflect.TypeOf(HealthView{}),
	}
	disallowed := []string{"admin", "shutdown", "retry", "delete", "mutate", "write"}
	for _, typ := range contractTypes {
		for _, field := range exportedJSONFieldNames(typ) {
			lower := strings.ToLower(field)
			for _, word := range disallowed {
				if strings.Contains(lower, word) {
					t.Fatalf("%s exposes disallowed write/admin field %q", typ.Name(), field)
				}
			}
		}
	}
}

func exportedJSONFieldNames(typ reflect.Type) []string {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	var fields []string
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			name = strings.Split(tag, ",")[0]
		}
		if name != "-" {
			fields = append(fields, name)
		}
	}
	return fields
}

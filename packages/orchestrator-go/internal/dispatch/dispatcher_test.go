package dispatch

import (
	"context"
	"strings"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/types"
)

func TestNativeDispatcherComposesExistingInvokeAgentSpec(t *testing.T) {
	result, err := NewNativeDispatcher().InvokeAgent(context.Background(), InvocationRequest{
		Agent: "builder",
		Task:  "ship the feature",
	})
	if err != nil {
		t.Fatalf("InvokeAgent returned error: %v", err)
	}
	if result.ExecutionMode != types.ExecutionNative {
		t.Fatalf("expected native execution mode, got %q", result.ExecutionMode)
	}
	if result.Spec == nil {
		t.Fatalf("expected composed spec")
	}
	if result.Spec.ID != "builder" || result.Spec.Base != "builder" || result.Spec.Model != "sonnet" {
		t.Fatalf("unexpected native spec identity/model: %+v", result.Spec)
	}
	if result.Spec.Prompt != "Task: ship the feature\nExecute as the builder agent." {
		t.Fatalf("native prompt shape changed: %q", result.Spec.Prompt)
	}
}

func TestDisabledA2ADispatcherReturnsClearNotConfiguredError(t *testing.T) {
	_, err := NewDisabledA2ADispatcher().InvokeAgent(context.Background(), InvocationRequest{Agent: "builder", Task: "ship"})
	if err == nil {
		t.Fatalf("expected disabled A2A error")
	}
	message := err.Error()
	for _, want := range []string{"A2A execution mode", "not configured", "Phase 1", "native or hybrid"} {
		if !strings.Contains(message, want) {
			t.Fatalf("A2A disabled error %q missing %q", message, want)
		}
	}
}

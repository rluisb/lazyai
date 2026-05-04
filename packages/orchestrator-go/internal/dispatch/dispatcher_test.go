package dispatch

import (
	"context"
	"strings"
	"testing"

	oconfig "github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/config"
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

func TestConfiguredA2ADispatcherReturnsPhase3NotImplementedForEnabledTarget(t *testing.T) {
	_, err := NewConfiguredDispatcher(types.ExecutionA2A, testA2AConfig()).InvokeAgent(context.Background(), InvocationRequest{
		Agent:   "builder",
		Task:    "ship",
		CliTool: string(types.HostOpenCode),
	})
	if err == nil {
		t.Fatalf("expected Phase 3 not implemented error")
	}
	message := err.Error()
	for _, want := range []string{"provider", "builder", "not implemented until Phase 3"} {
		if !strings.Contains(message, want) {
			t.Fatalf("configured A2A error %q missing %q", message, want)
		}
	}
}

func TestConfiguredA2ADispatcherRejectsCliToolMismatch(t *testing.T) {
	_, err := NewConfiguredDispatcher(types.ExecutionA2A, testA2AConfig()).InvokeAgent(context.Background(), InvocationRequest{
		Agent:   "builder",
		Task:    "ship",
		CliTool: string(types.HostCopilot),
	})
	if err == nil || !strings.Contains(err.Error(), "cliTool") {
		t.Fatalf("expected cliTool mismatch error, got %v", err)
	}
}

func TestConfiguredHybridDispatcherFallsBackToNative(t *testing.T) {
	result, err := NewConfiguredDispatcher(types.ExecutionHybrid, testA2AConfig()).InvokeAgent(context.Background(), InvocationRequest{
		Agent:   "builder",
		Task:    "ship the feature",
		CliTool: string(types.HostOpenCode),
	})
	if err != nil {
		t.Fatalf("hybrid InvokeAgent returned error: %v", err)
	}
	if result.ExecutionMode != types.ExecutionNative || result.Spec == nil {
		t.Fatalf("expected native fallback result, got %+v", result)
	}
	if result.Spec.Prompt != "Task: ship the feature\nExecute as the builder agent." {
		t.Fatalf("hybrid fallback changed native response shape: %+v", result.Spec)
	}
}

func testA2AConfig() *oconfig.Config {
	enabled := true
	return &oconfig.Config{
		Version:   1,
		Execution: oconfig.ExecutionConfig{Mode: types.ExecutionA2A},
		Providers: map[string]oconfig.ProviderConfig{
			"local": {Endpoint: "https://a2a.example.test/rpc", Auth: oconfig.AuthConfig{Type: oconfig.AuthNone}},
		},
		Agents: map[string]oconfig.AgentConfig{
			"builder": {Provider: "local", Enabled: &enabled, Tools: []string{string(types.HostOpenCode)}},
		},
	}
}

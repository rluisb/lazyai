package adapter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestRegistryGet_UnsupportedToolListsRegisteredToolsDeterministically(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get(types.ToolId("gemini"))
	if err == nil {
		t.Fatal("expected unsupported-tool error")
	}
	supported := r.List()
	wantSupported := []types.ToolId{
		types.ToolIdAntigravity,
		types.ToolIdClaudeCode,
		types.ToolIdCopilot,
		types.ToolIdKiro,
		types.ToolIdOpenCode,
		types.ToolIdOmp,
		types.ToolIdPi,
	}

	if len(supported) != len(wantSupported) {
		t.Fatalf("supported tools count = %d, want %d", len(supported), len(wantSupported))
	}
	for i := range wantSupported {
		if supported[i] != wantSupported[i] {
			t.Fatalf("supported tool %d = %q, want %q", i, supported[i], wantSupported[i])
		}
	}

	supportedStrings := make([]string, len(supported))
	for i, id := range supported {
		supportedStrings[i] = string(id)
	}

	wantErr := fmt.Sprintf("unsupported tool %q (supported tools: %s)", types.ToolId("gemini"), strings.Join(supportedStrings, ", "))
	if err.Error() != wantErr {
		t.Fatalf("error = %q, want %q", err.Error(), wantErr)
	}
}

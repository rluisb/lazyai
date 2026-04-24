package adapter

import (
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestAllAdapters_SatisfyToolAdapter(t *testing.T) {
	var _ ToolAdapter = (*OpenCodeAdapter)(nil)
}

func TestCanRunHeadless_Values(t *testing.T) {
	adapter := &OpenCodeAdapter{}
	if adapter.CanRunHeadless() {
		t.Errorf("OpenCode.CanRunHeadless() = true, want false")
	}
}

func TestRunHeadlessValidation_NoOpAdapters(t *testing.T) {
	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
	}

	if err := (&OpenCodeAdapter{}).RunHeadlessValidation(ctx); err != nil {
		t.Errorf("OpenCode.RunHeadlessValidation() returned error: %v", err)
	}
}

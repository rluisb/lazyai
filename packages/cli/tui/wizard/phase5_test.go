package wizard

import (
	"reflect"
	"testing"
)

func TestBuildPhase5Result(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args Phase5Result
		want Phase5Result
	}{
		{
			name: "all optional integrations omitted keeps explicit paths except memory fallback",
			args: Phase5Result{},
			want: Phase5Result{MemoryPath: ".specify/memory"},
		},
		{
			name: "obsidian enabled preserves provided vault path",
			args: Phase5Result{MemoryPath: "memory/custom", EnableObsidian: true, ObsidianVaultPath: "/vault"},
			want: Phase5Result{MemoryPath: "memory/custom", EnableObsidian: true, ObsidianVaultPath: "/vault"},
		},
		{
			name: "codegraph enabled fills default data path when empty",
			args: Phase5Result{EnableCodegraph: true},
			want: Phase5Result{MemoryPath: ".specify/memory", EnableCodegraph: true, CodegraphDataPath: ".codegraph/"},
		},
		{
			name: "codegraph enabled preserves provided data path",
			args: Phase5Result{EnableCodegraph: true, CodegraphDataPath: "custom-graph"},
			want: Phase5Result{MemoryPath: ".specify/memory", EnableCodegraph: true, CodegraphDataPath: "custom-graph"},
		},
		{
			name: "all enabled applies default fallbacks",
			args: Phase5Result{EnableObsidian: true, ObsidianVaultPath: "/vault", EnableCodegraph: true},
			want: Phase5Result{MemoryPath: ".specify/memory", EnableObsidian: true, ObsidianVaultPath: "/vault", EnableCodegraph: true, CodegraphDataPath: ".codegraph/"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildPhase5Result(
				tt.args.MemoryPath,
				tt.args.EnableObsidian,
				tt.args.ObsidianVaultPath,
				tt.args.EnableCodegraph,
				tt.args.CodegraphDataPath,
			)

			if !reflect.DeepEqual(*got, tt.want) {
				t.Fatalf("buildPhase5Result() = %#v, want %#v", *got, tt.want)
			}
		})
	}
}

func TestPhase5StepInfoTitles(t *testing.T) {
	t.Parallel()

	if got, want := phase5MemoryPathStepInfo().Title(), "Optional Tooling — 1/2: Memory Path"; got != want {
		t.Fatalf("phase5MemoryPathStepInfo().Title() = %q, want %q", got, want)
	}
}

func TestDefaultPhase5ResultEnablesToolingByDefault(t *testing.T) {
	t.Parallel()

	result := defaultPhase5Result()
	if !result.EnableObsidian || !result.EnableCodegraph {
		t.Fatalf("tooling defaults not enabled: %#v", result)
	}
	if result.MemoryPath != ".specify/memory" || result.CodegraphDataPath != ".codegraph/" {
		t.Fatalf("unexpected phase 5 defaults: %#v", result)
	}
}

func TestRunPhase5NonInteractiveDefaults(t *testing.T) {
	t.Parallel()

	result, action, err := RunPhase5(nil, true)
	if err != nil {
		t.Fatalf("RunPhase5: %v", err)
	}
	if action != PhaseContinue {
		t.Fatalf("action = %v, want %v", action, PhaseContinue)
	}
	if result.MemoryPath != ".specify/memory" {
		t.Fatalf("MemoryPath = %q, want .specify/memory", result.MemoryPath)
	}
	if !result.EnableObsidian || !result.EnableCodegraph {
		t.Fatalf("tooling defaults not enabled: %#v", result)
	}
	if result.CodegraphDataPath != ".codegraph/" {
		t.Fatalf("CodegraphDataPath = %q, want .codegraph/", result.CodegraphDataPath)
	}
}

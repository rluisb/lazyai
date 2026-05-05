package wizard

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestOpenCodePluginsExactURLs(t *testing.T) {
	t.Parallel()

	want := []string{
		"https://github.com/Opencode-DCP/opencode-dynamic-context-pruning",
		"https://github.com/spoons-and-mirrors/subtask2",
		"https://github.com/JRedeker/opencode-shell-strategy",
		"https://github.com/boxpositron/envsitter-guard",
		"https://github.com/kdcokenny/opencode-background-agents",
	}

	if !reflect.DeepEqual(opencodePluginURLs, want) {
		t.Fatalf("opencodePluginURLs = %#v, want %#v", opencodePluginURLs, want)
	}

	data, err := os.ReadFile("../../library/opencode/plugins.json")
	if err != nil {
		t.Fatalf("read OpenCode plugin catalog: %v", err)
	}
	var catalog []struct {
		Module string `json:"module"`
	}
	if err := json.Unmarshal(data, &catalog); err != nil {
		t.Fatalf("parse OpenCode plugin catalog: %v", err)
	}
	got := make([]string, 0, len(catalog))
	for _, plugin := range catalog {
		got = append(got, plugin.Module)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("OpenCode plugin catalog = %#v, want %#v", got, want)
	}
}

func TestBuildPhase5Result(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args Phase5Result
		want Phase5Result
	}{
		{
			name: "all optional integrations disabled keeps explicit paths except memory fallback",
			args: Phase5Result{},
			want: Phase5Result{MemoryPath: ".specify/memory"},
		},
		{
			name: "obsidian enabled preserves provided vault path",
			args: Phase5Result{MemoryPath: "memory/custom", EnableObsidian: true, ObsidianVaultPath: "/vault"},
			want: Phase5Result{MemoryPath: "memory/custom", EnableObsidian: true, ObsidianVaultPath: "/vault"},
		},
		{
			name: "qmd enabled keeps empty index path for global qmd index",
			args: Phase5Result{EnableQmd: true},
			want: Phase5Result{MemoryPath: ".specify/memory", EnableQmd: true},
		},
		{
			name: "qmd enabled preserves provided index path",
			args: Phase5Result{EnableQmd: true, QmdIndexPath: "custom-index"},
			want: Phase5Result{MemoryPath: ".specify/memory", EnableQmd: true, QmdIndexPath: "custom-index"},
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
			name: "graphify enabled fills default data path when empty",
			args: Phase5Result{EnableGraphify: true},
			want: Phase5Result{MemoryPath: ".specify/memory", EnableGraphify: true, GraphifyDataPath: "graphify-out"},
		},
		{
			name: "graphify enabled preserves provided data path",
			args: Phase5Result{EnableGraphify: true, GraphifyDataPath: "custom-graphify"},
			want: Phase5Result{MemoryPath: ".specify/memory", EnableGraphify: true, GraphifyDataPath: "custom-graphify"},
		},
		{
			name: "all enabled applies default fallbacks",
			args: Phase5Result{EnableObsidian: true, ObsidianVaultPath: "/vault", EnableQmd: true, EnableCodegraph: true, EnableGraphify: true},
			want: Phase5Result{MemoryPath: ".specify/memory", EnableObsidian: true, ObsidianVaultPath: "/vault", EnableQmd: true, EnableCodegraph: true, CodegraphDataPath: ".codegraph/", EnableGraphify: true, GraphifyDataPath: "graphify-out"},
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
				tt.args.EnableQmd,
				tt.args.QmdIndexPath,
				tt.args.EnableCodegraph,
				tt.args.CodegraphDataPath,
				tt.args.EnableGraphify,
				tt.args.GraphifyDataPath,
				tt.args.OpenCodePlugins,
			)

			if !reflect.DeepEqual(*got, tt.want) {
				t.Fatalf("buildPhase5Result() = %#v, want %#v", *got, tt.want)
			}
		})
	}
}

func TestPhase5StepInfoTitles(t *testing.T) {
	t.Parallel()

	state := Phase5Result{}
	if got, want := phase5MemoryPathStepInfo(state).Title(), "Optional Tooling — 1/1: Memory Path"; got != want {
		t.Fatalf("phase5MemoryPathStepInfo().Title() = %q, want %q", got, want)
	}
	if got, want := phase5MemoryPathStepInfo(state, true).Title(), "Optional Tooling — 1/2: Memory Path"; got != want {
		t.Fatalf("phase5MemoryPathStepInfo(showPlugins).Title() = %q, want %q", got, want)
	}
	if got, want := phase5OpenCodePluginsStepInfo(state, true).Title(), "Optional Tooling — 2/2: OpenCode Plugins"; got != want {
		t.Fatalf("phase5OpenCodePluginsStepInfo().Title() = %q, want %q", got, want)
	}
}

func TestDefaultPhase5ResultEnablesToolingByDefault(t *testing.T) {
	t.Parallel()

	result := defaultPhase5Result()
	if !result.EnableObsidian || !result.EnableQmd || !result.EnableCodegraph || !result.EnableGraphify {
		t.Fatalf("tooling defaults not enabled: %#v", result)
	}
	if result.MemoryPath != ".specify/memory" || result.QmdIndexPath != "" || result.CodegraphDataPath != ".codegraph/" || result.GraphifyDataPath != "graphify-out" {
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
	if !result.EnableObsidian || !result.EnableQmd || !result.EnableCodegraph || !result.EnableGraphify {
		t.Fatalf("tooling defaults not enabled: %#v", result)
	}
	if result.QmdIndexPath != "" {
		t.Fatalf("QmdIndexPath = %q, want empty", result.QmdIndexPath)
	}
	if result.CodegraphDataPath != ".codegraph/" {
		t.Fatalf("CodegraphDataPath = %q, want .codegraph/", result.CodegraphDataPath)
	}
	if result.GraphifyDataPath != "graphify-out" {
		t.Fatalf("GraphifyDataPath = %q, want graphify-out", result.GraphifyDataPath)
	}
}

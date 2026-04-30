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
			want: Phase5Result{MemoryPath: "specs/memory"},
		},
		{
			name: "obsidian enabled preserves provided vault path",
			args: Phase5Result{MemoryPath: "memory/custom", EnableObsidian: true, ObsidianVaultPath: "/vault"},
			want: Phase5Result{MemoryPath: "memory/custom", EnableObsidian: true, ObsidianVaultPath: "/vault"},
		},
		{
			name: "qmd enabled fills default index path when empty",
			args: Phase5Result{EnableQmd: true},
			want: Phase5Result{MemoryPath: "specs/memory", EnableQmd: true, QmdIndexPath: ".qmd-index"},
		},
		{
			name: "qmd enabled preserves provided index path",
			args: Phase5Result{EnableQmd: true, QmdIndexPath: "custom-index"},
			want: Phase5Result{MemoryPath: "specs/memory", EnableQmd: true, QmdIndexPath: "custom-index"},
		},
		{
			name: "codegraph enabled fills default data path when empty",
			args: Phase5Result{EnableCodegraph: true},
			want: Phase5Result{MemoryPath: "specs/memory", EnableCodegraph: true, CodegraphDataPath: ".codegraph"},
		},
		{
			name: "codegraph enabled preserves provided data path",
			args: Phase5Result{EnableCodegraph: true, CodegraphDataPath: "custom-graph"},
			want: Phase5Result{MemoryPath: "specs/memory", EnableCodegraph: true, CodegraphDataPath: "custom-graph"},
		},
		{
			name: "all enabled applies both default fallbacks",
			args: Phase5Result{EnableObsidian: true, ObsidianVaultPath: "/vault", EnableQmd: true, EnableCodegraph: true},
			want: Phase5Result{MemoryPath: "specs/memory", EnableObsidian: true, ObsidianVaultPath: "/vault", EnableQmd: true, QmdIndexPath: ".qmd-index", EnableCodegraph: true, CodegraphDataPath: ".codegraph"},
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

	allDisabled := Phase5Result{}
	if got, want := phase5EnableQmdStepInfo(allDisabled).Title(), "Optional Tooling — 3/4: Enable qmd"; got != want {
		t.Fatalf("phase5EnableQmdStepInfo().Title() = %q, want %q", got, want)
	}
	if got, want := phase5EnableCodegraphStepInfo(allDisabled).Title(), "Optional Tooling — 4/4: Enable Codegraph"; got != want {
		t.Fatalf("phase5EnableCodegraphStepInfo().Title() = %q, want %q", got, want)
	}

	obsidianOnly := Phase5Result{EnableObsidian: true}
	if got, want := phase5EnableQmdStepInfo(obsidianOnly).Title(), "Optional Tooling — 4/5: Enable qmd"; got != want {
		t.Fatalf("phase5EnableQmdStepInfo().Title() = %q, want %q", got, want)
	}

	allEnabled := Phase5Result{EnableObsidian: true, EnableQmd: true, EnableCodegraph: true}
	if got, want := phase5CodegraphDataPathStepInfo(allEnabled, false).Title(), "Optional Tooling — 7/7: Codegraph Data Path"; got != want {
		t.Fatalf("phase5CodegraphDataPathStepInfo().Title() = %q, want %q", got, want)
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
	if result.MemoryPath != "specs/memory" {
		t.Fatalf("MemoryPath = %q, want specs/memory", result.MemoryPath)
	}
	if result.QmdIndexPath != ".qmd-index" {
		t.Fatalf("QmdIndexPath = %q, want .qmd-index", result.QmdIndexPath)
	}
	if result.CodegraphDataPath != ".codegraph" {
		t.Fatalf("CodegraphDataPath = %q, want .codegraph", result.CodegraphDataPath)
	}
}

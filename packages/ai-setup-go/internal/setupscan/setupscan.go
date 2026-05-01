package setupscan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/adapter"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/jsonc"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

type Options struct {
	HomeDir   string
	TargetDir string
	Adopt     bool
	Import    bool
}

type Inventory struct {
	CurrentState CurrentState     `json:"currentState"`
	DesiredState DesiredState     `json:"desiredState"`
	Operation    *OperationResult `json:"operation,omitempty"`
}

type CurrentState struct {
	SharedPaths []ObservedPath   `json:"sharedPaths"`
	Targets     []ObservedTarget `json:"targets"`
	Agents      []ObservedAgent  `json:"agents,omitempty"`
}

type DesiredState struct {
	SharedPaths []DesiredPath   `json:"sharedPaths"`
	Targets     []DesiredTarget `json:"targets"`
}

type ObservedPath struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
}

type DesiredPath struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

type ObservedTarget struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Detections []TargetDetection `json:"detections"`
}

type TargetDetection struct {
	Scope         string     `json:"scope"`
	Origin        string     `json:"origin"`
	RootPath      string     `json:"rootPath"`
	Status        string     `json:"status"`
	State         string     `json:"state"`
	Version       string     `json:"version"`
	ObservedFiles []string   `json:"observedFiles"`
	Reasons       []string   `json:"reasons,omitempty"`
	MCPEntries    []MCPEntry `json:"mcpEntries,omitempty"`
}

type MCPEntry struct {
	Name       string   `json:"name"`
	ConfigPath string   `json:"configPath"`
	State      string   `json:"state"`
	Reasons    []string `json:"reasons,omitempty"`
}

type DesiredTarget struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	SupportedScopes []string      `json:"supportedScopes"`
	CandidateRoots  []DesiredRoot `json:"candidateRoots"`
}

type DesiredRoot struct {
	Scope         string   `json:"scope"`
	Origin        string   `json:"origin"`
	RootPath      string   `json:"rootPath"`
	ExpectedFiles []string `json:"expectedFiles"`
}

type targetSpec struct {
	Tool            types.ToolId
	Name            string
	SupportedScopes []types.SetupScope
	Roots           []rootSpec
}

type rootSpec struct {
	Scope         types.SetupScope
	Origin        string
	Resolve       func(opts Options) (string, error)
	CountRootOnly bool
	ExpectedFiles []string
	OptionalPaths []string
	VersionFiles  []string
}

var tomlVersionPattern = regexp.MustCompile(`(?m)^version\s*=\s*"([^"]+)"`)

func Scan(opts Options) (*Inventory, error) {
	if opts.TargetDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("determine working directory: %w", err)
		}
		opts.TargetDir = wd
	}
	if opts.HomeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("determine home directory: %w", err)
		}
		opts.HomeDir = home
	}

	specs := supportedTargets(opts)
	registry, err := loadRegistry(aiSetupHome(opts))
	if err != nil {
		return nil, err
	}
	inventory := &Inventory{
		CurrentState: CurrentState{
			SharedPaths: observedSharedPaths(opts),
			Targets:     observedTargets(specs, opts, registry),
			Agents:      observedAgents(opts),
		},
		DesiredState: DesiredState{
			SharedPaths: desiredSharedPaths(opts),
			Targets:     desiredTargets(specs, opts),
		},
	}

	return inventory, nil
}

func observedSharedPaths(opts Options) []ObservedPath {
	paths := []ObservedPath{
		{ID: "global-ai-setup", Path: filepath.Join(opts.HomeDir, ".ai-setup")},
		{ID: "project-ai", Path: filepath.Join(opts.TargetDir, ".ai")},
	}
	for i := range paths {
		paths[i].Exists = files.DirExists(paths[i].Path)
	}
	return paths
}

func desiredSharedPaths(opts Options) []DesiredPath {
	return []DesiredPath{
		{ID: "global-ai-setup", Description: "Global ai-setup managed directory", Path: filepath.Join(opts.HomeDir, ".ai-setup")},
		{ID: "project-ai", Description: "Project-local ai-setup directory", Path: filepath.Join(opts.TargetDir, ".ai")},
	}
}

func observedTargets(specs []targetSpec, opts Options, registry *scanRegistry) []ObservedTarget {
	targets := make([]ObservedTarget, 0, len(specs))
	for _, spec := range specs {
		detections := make([]TargetDetection, 0, len(spec.Roots))
		for _, root := range spec.Roots {
			rootPath, err := root.Resolve(opts)
			if err != nil || rootPath == "" {
				continue
			}
			observedFiles := collectObservedFiles(rootPath, append(root.ExpectedFiles, root.OptionalPaths...))
			status := "missing"
			if len(observedFiles) > 0 || (root.CountRootOnly && files.DirExists(rootPath)) {
				status = "detected"
			}
			detection := TargetDetection{
				Scope:         string(root.Scope),
				Origin:        root.Origin,
				RootPath:      rootPath,
				Status:        status,
				Version:       detectVersion(rootPath, root.VersionFiles),
				ObservedFiles: observedFiles,
			}
			applyDetectionState(&detection, spec.Tool, registry)
			detections = append(detections, detection)
		}
		sort.Slice(detections, func(i, j int) bool {
			if detections[i].Origin != detections[j].Origin {
				return detections[i].Origin < detections[j].Origin
			}
			if detections[i].Scope != detections[j].Scope {
				return detections[i].Scope < detections[j].Scope
			}
			return detections[i].RootPath < detections[j].RootPath
		})
		targets = append(targets, ObservedTarget{
			ID:         string(spec.Tool),
			Name:       spec.Name,
			Detections: detections,
		})
	}
	return targets
}

func desiredTargets(specs []targetSpec, opts Options) []DesiredTarget {
	targets := make([]DesiredTarget, 0, len(specs))
	for _, spec := range specs {
		candidateRoots := make([]DesiredRoot, 0, len(spec.Roots))
		for _, root := range spec.Roots {
			rootPath, err := root.Resolve(opts)
			if err != nil || rootPath == "" {
				continue
			}
			expectedFiles := append([]string{}, root.ExpectedFiles...)
			expectedFiles = append(expectedFiles, root.OptionalPaths...)
			sort.Strings(expectedFiles)
			candidateRoots = append(candidateRoots, DesiredRoot{
				Scope:         string(root.Scope),
				Origin:        root.Origin,
				RootPath:      rootPath,
				ExpectedFiles: expectedFiles,
			})
		}
		supportedScopes := make([]string, 0, len(spec.SupportedScopes))
		for _, scope := range spec.SupportedScopes {
			supportedScopes = append(supportedScopes, string(scope))
		}
		targets = append(targets, DesiredTarget{
			ID:              string(spec.Tool),
			Name:            spec.Name,
			SupportedScopes: supportedScopes,
			CandidateRoots:  candidateRoots,
		})
	}
	return targets
}

func collectObservedFiles(rootPath string, candidates []string) []string {
	set := make(map[string]bool)
	for _, relativePath := range candidates {
		if relativePath == "" {
			continue
		}
		absPath := filepath.Join(rootPath, relativePath)
		if files.FileExists(absPath) || files.DirExists(absPath) {
			set[filepath.ToSlash(relativePath)] = true
		}
	}
	observed := make([]string, 0, len(set))
	for file := range set {
		observed = append(observed, file)
	}
	sort.Strings(observed)
	return observed
}

func detectVersion(rootPath string, candidates []string) string {
	for _, relativePath := range candidates {
		version := detectVersionFromFile(filepath.Join(rootPath, relativePath))
		if version != "unknown" {
			return version
		}
	}
	return "unknown"
}

func detectVersionFromFile(path string) string {
	if !files.FileExists(path) {
		return "unknown"
	}
	if strings.HasSuffix(path, ".jsonc") {
		parsed, err := jsonc.ReadJSONCFile(path)
		if err == nil {
			if version := versionFromMap(parsed); version != "" {
				return version
			}
		}
		return "unknown"
	}
	if strings.HasSuffix(path, ".json") {
		data, err := os.ReadFile(path)
		if err == nil {
			var parsed map[string]any
			if err := json.Unmarshal(data, &parsed); err == nil {
				if version := versionFromMap(parsed); version != "" {
					return version
				}
			}
		}
		return "unknown"
	}
	if strings.HasSuffix(path, ".toml") {
		data, err := os.ReadFile(path)
		if err == nil {
			matches := tomlVersionPattern.FindStringSubmatch(string(data))
			if len(matches) == 2 {
				return matches[1]
			}
		}
	}
	return "unknown"
}

func versionFromMap(parsed map[string]any) string {
	if parsed == nil {
		return ""
	}
	if version, ok := parsed["version"]; ok {
		switch v := version.(type) {
		case string:
			if v != "" {
				return v
			}
		case float64:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func supportedTargets(opts Options) []targetSpec {
	registry := adapter.NewRegistry()
	toolIDs := registry.List()
	sort.Slice(toolIDs, func(i, j int) bool { return toolIDs[i] < toolIDs[j] })

	specs := make([]targetSpec, 0, len(toolIDs))
	for _, toolID := range toolIDs {
		adapterInstance, err := registry.Get(toolID)
		if err != nil {
			continue
		}
		specs = append(specs, targetSpec{
			Tool:            toolID,
			Name:            adapterInstance.Name(),
			SupportedScopes: supportedScopesForTool(toolID),
			Roots:           rootsForTool(toolID, opts),
		})
	}
	return specs
}

func supportedScopesForTool(tool types.ToolId) []types.SetupScope {
	return []types.SetupScope{types.SetupScopeGlobal, types.SetupScopeProject, types.SetupScopeWorkspace}
}

func rootsForTool(tool types.ToolId, opts Options) []rootSpec {
	switch tool {
	case types.ToolIdClaudeCode:
		return []rootSpec{
			toolRootSpec(tool, types.SetupScopeGlobal, "global", opts, []string{"settings.json"}, []string{"settings.local.json", "agents", "skills", "commands", "output-styles"}),
			toolRootSpec(tool, types.SetupScopeProject, "project", opts, []string{"settings.json"}, []string{"settings.local.json", "agents", "skills", "commands", "output-styles"}),
			toolRootSpec(tool, types.SetupScopeWorkspace, "workspace", opts, []string{"settings.json"}, []string{"settings.local.json", "agents", "skills", "commands", "output-styles"}),
		}
	case types.ToolIdCopilot:
		return []rootSpec{
			copilotGlobalRootSpec(opts),
			toolRootSpec(tool, types.SetupScopeProject, "project", opts, []string{"copilot-instructions.md"}, []string{"agents", "instructions", "prompts", "chatmodes"}),
			toolRootSpec(tool, types.SetupScopeWorkspace, "workspace", opts, []string{"copilot-instructions.md"}, []string{"agents", "instructions", "prompts", "chatmodes"}),
		}
	case types.ToolIdOpenCode:
		return []rootSpec{
			toolRootSpec(tool, types.SetupScopeGlobal, "global", opts, []string{"opencode.jsonc"}, []string{"opencode.json", "agents", "skills", "commands", "modes", "AGENTS.md"}),
			toolRootSpec(tool, types.SetupScopeProject, "project", opts, []string{"opencode.jsonc"}, []string{"opencode.json", "agents", "skills", "commands", "modes", "AGENTS.md"}),
			toolRootSpec(tool, types.SetupScopeWorkspace, "workspace", opts, []string{"opencode.jsonc"}, []string{"opencode.json", "agents", "skills", "commands", "modes", "AGENTS.md"}),
		}
	}
	return nil
}

func toolRootSpec(tool types.ToolId, scope types.SetupScope, origin string, opts Options, expectedFiles, optionalPaths []string) rootSpec {
	ctx := &adapter.AdapterContext{TargetDir: opts.TargetDir, HomeDir: opts.HomeDir, SetupScope: scope}
	return rootSpec{
		Scope:         scope,
		Origin:        origin,
		Resolve:       func(_ Options) (string, error) { return adapter.ResolveToolRoot(tool, scope, ctx) },
		CountRootOnly: false,
		ExpectedFiles: expectedFiles,
		OptionalPaths: optionalPaths,
		VersionFiles:  expectedFiles,
	}
}

func copilotGlobalRootSpec(opts Options) rootSpec {
	ctx := &adapter.AdapterContext{TargetDir: opts.TargetDir, HomeDir: opts.HomeDir, SetupScope: types.SetupScopeGlobal}
	return rootSpec{
		Scope:  types.SetupScopeGlobal,
		Origin: "global",
		Resolve: func(_ Options) (string, error) {
			return adapter.ResolveToolRoot(types.ToolIdCopilot, types.SetupScopeGlobal, ctx)
		},
		CountRootOnly: true,
		ExpectedFiles: []string{"mcp-config.json"},
		OptionalPaths: []string{},
		VersionFiles:  []string{"mcp-config.json"},
	}
}

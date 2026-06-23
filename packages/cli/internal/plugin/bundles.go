package plugin

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

type BundleTarget string

const (
	BundleTargetClaude     BundleTarget = "claude"
	BundleTargetCopilotCLI BundleTarget = "copilot-cli"
	BundleTargetOmp        BundleTarget = "omp"
	BundleTargetPi         BundleTarget = "pi"
)

func NormalizeTarget(raw string) (BundleTarget, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "", "claude", "claude-code":
		return BundleTargetClaude, nil
	case "copilot", "copilot-cli":
		return BundleTargetCopilotCLI, nil
	case "omp":
		return BundleTargetOmp, nil
	case "pi":
		return BundleTargetPi, nil
	default:
		return "", fmt.Errorf("unsupported bundle target %q (supported: claude, copilot-cli, omp, pi)", raw)
	}
}

func BuildTarget(libFS fs.FS, outDir, version string, target BundleTarget) (BuildResult, error) {
	switch target {
	case BundleTargetClaude:
		return Build(libFS, outDir, version)
	case BundleTargetCopilotCLI:
		return buildCopilotCLI(libFS, outDir, version)
	case BundleTargetOmp:
		return buildOmpBundle(libFS, outDir)
	case BundleTargetPi:
		return buildPiBundle(libFS, outDir)
	default:
		return BuildResult{}, fmt.Errorf("unsupported bundle target %q", target)
	}
}

func buildCopilotCLI(libFS fs.FS, outDir, version string) (BuildResult, error) {
	stageDir, cleanup, err := buildStage(libFS, types.ToolIdCopilot, false)
	if err != nil {
		return BuildResult{}, err
	}
	defer cleanup()

	if err := writeSimplePluginManifest(filepath.Join(outDir, "plugin.json"), version); err != nil {
		return BuildResult{}, err
	}
	if err := copyDirIfExists(filepath.Join(stageDir, ".github", "agents"), filepath.Join(outDir, "agents")); err != nil {
		return BuildResult{}, err
	}
	if err := copyDirIfExists(filepath.Join(stageDir, ".github", "skills"), filepath.Join(outDir, "skills")); err != nil {
		return BuildResult{}, err
	}
	if err := copyCopilotHooks(filepath.Join(stageDir, ".github", "hooks"), outDir); err != nil {
		return BuildResult{}, err
	}
	if err := copyLibraryFileIfExists(libFS, "mcp/catalog.json", filepath.Join(outDir, ".mcp.json"), 0o644); err != nil {
		return BuildResult{}, err
	}
	count, err := countFiles(outDir)
	if err != nil {
		return BuildResult{}, err
	}
	return BuildResult{OutDir: outDir, FileCount: count}, nil
}

func buildOmpBundle(libFS fs.FS, outDir string) (BuildResult, error) {
	stageDir, cleanup, err := buildStage(libFS, types.ToolIdOmp, true)
	if err != nil {
		return BuildResult{}, err
	}
	defer cleanup()

	ompDir := filepath.Join(stageDir, ".omp")
	for _, part := range []string{"skills", "commands", "hooks"} {
		if err := copyDirIfExists(filepath.Join(ompDir, part), filepath.Join(outDir, part)); err != nil {
			return BuildResult{}, err
		}
	}
	if err := copyFileIfExists(filepath.Join(ompDir, "mcp.json"), filepath.Join(outDir, "mcp.json")); err != nil {
		return BuildResult{}, err
	}
	if !files.FileExists(filepath.Join(outDir, "mcp.json")) {
		if err := copyLibraryFileIfExists(libFS, "mcp/catalog.json", filepath.Join(outDir, "mcp.json"), 0o644); err != nil {
			return BuildResult{}, err
		}
	}
	count, err := countFiles(outDir)
	if err != nil {
		return BuildResult{}, err
	}
	return BuildResult{OutDir: outDir, FileCount: count}, nil
}

func buildPiBundle(libFS fs.FS, outDir string) (BuildResult, error) {
	stageDir, cleanup, err := buildStage(libFS, types.ToolIdPi, false)
	if err != nil {
		return BuildResult{}, err
	}
	defer cleanup()

	piDir := filepath.Join(stageDir, ".pi")
	for _, part := range []string{"agents", "skills", "prompts", "extensions"} {
		if err := copyDirIfExists(filepath.Join(piDir, part), filepath.Join(outDir, part)); err != nil {
			return BuildResult{}, err
		}
	}
	count, err := countFiles(outDir)
	if err != nil {
		return BuildResult{}, err
	}
	return BuildResult{OutDir: outDir, FileCount: count}, nil
}

func buildStage(libFS fs.FS, tool types.ToolId, seedMCP bool) (string, func(), error) {
	if libFS == nil {
		return "", nil, fmt.Errorf("plugin.BuildTarget: libFS is nil")
	}
	stageDir, err := os.MkdirTemp("", "lazyai-bundle-stage-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(stageDir) }
	selections, err := availableSelections(libFS)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	ctx := &adapter.AdapterContext{
		TargetDir:  stageDir,
		HomeDir:    stageDir,
		SetupScope: types.SetupScopeProject,
		LibraryFS:  libFS,
		Strategy:   types.ConflictStrategyAlign,
		Selections: selections,
	}
	ad, err := adapterInstance(tool)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	if _, err := ad.Install(ctx); err != nil {
		cleanup()
		return "", nil, err
	}
	if seedMCP {
		if err := seedCanonicalMCP(libFS, stageDir); err != nil {
			cleanup()
			return "", nil, err
		}
		if _, err := ad.CompileMCP(adapter.CompileContext{TargetDir: stageDir, HomeDir: stageDir, SetupScope: types.SetupScopeProject}); err != nil {
			cleanup()
			return "", nil, err
		}
	}
	return stageDir, cleanup, nil
}

func availableSelections(libFS fs.FS) (adapter.AdapterSelections, error) {
	var out adapter.AdapterSelections
	if ids, err := listIDs(libFS, "canonical/agents"); err != nil {
		return out, err
	} else {
		out.Agents = make([]types.AgentId, 0, len(ids))
		for _, id := range ids {
			out.Agents = append(out.Agents, types.AgentId(id))
		}
	}
	if ids, err := listIDs(libFS, "skills"); err != nil {
		return out, err
	} else {
		out.Skills = make([]types.SkillId, 0, len(ids))
		for _, id := range ids {
			out.Skills = append(out.Skills, types.SkillId(id))
		}
	}
	if ids, err := listIDs(libFS, "prompts"); err != nil {
		return out, err
	} else {
		out.Prompts = make([]types.PromptId, 0, len(ids))
		for _, id := range ids {
			out.Prompts = append(out.Prompts, types.PromptId(id))
		}
	}
	if ids, err := listIDs(libFS, "chatmodes"); err != nil {
		return out, err
	} else {
		out.ChatModes = make([]types.ChatModeId, 0, len(ids))
		for _, id := range ids {
			out.ChatModes = append(out.ChatModes, types.ChatModeId(id))
		}
	}
	return out, nil
}

func listIDs(libFS fs.FS, dir string) ([]string, error) {
	entries, err := fs.ReadDir(libFS, dir)
	if err != nil {
		if errorsIsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		switch dir {
		case "chatmodes":
			name = strings.TrimSuffix(strings.TrimSuffix(name, filepath.Ext(name)), ".chatmode")
		default:
			name = strings.TrimSuffix(name, filepath.Ext(name))
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}

func adapterInstance(tool types.ToolId) (adapter.ToolAdapter, error) {
	switch tool {
	case types.ToolIdCopilot:
		return &adapter.CopilotAdapter{}, nil
	case types.ToolIdOmp:
		return &adapter.OmpAdapter{}, nil
	case types.ToolIdPi:
		return &adapter.PiAdapter{}, nil
	default:
		return nil, fmt.Errorf("unsupported bundle tool %q", tool)
	}
}

func seedCanonicalMCP(libFS fs.FS, targetDir string) error {
	data, err := fs.ReadFile(libFS, "mcp/catalog.json")
	if err != nil {
		if os.IsNotExist(err) || err == fs.ErrNotExist {
			return nil
		}
		return err
	}
	path := filepath.Join(targetDir, ".ai", "mcp.json")
	if err := files.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	return files.SafeWriteFile(path, data, 0o644)
}

func writeSimplePluginManifest(path, version string) error {
	manifest := map[string]any{
		"name":        "lazyai-vibelab",
		"version":     version,
		"description": "vibe-lab agents, skills, hooks, and MCP configuration compiled by LazyAI.",
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := files.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	return files.SafeWriteFile(path, data, 0o644)
}

func copyCopilotHooks(srcDir, outDir string) error {
	if !files.DirExists(srcDir) {
		return nil
	}
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	merged := map[string]any{"version": 1, "hooks": map[string]any{}}
	hooksMap := merged["hooks"].(map[string]any)
	for _, entry := range entries {
		name := entry.Name()
		srcPath := filepath.Join(srcDir, name)
		switch {
		case strings.HasSuffix(name, ".sh"):
			if err := copyFileIfExists(srcPath, filepath.Join(outDir, "hooks", name)); err != nil {
				return err
			}
		case strings.HasSuffix(name, ".json"):
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			var payload map[string]any
			if err := json.Unmarshal(data, &payload); err != nil {
				return fmt.Errorf("parse hook config %s: %w", name, err)
			}
			rawHooks, _ := payload["hooks"].(map[string]any)
			for event, rawList := range rawHooks {
				items, _ := rawList.([]any)
				for _, item := range items {
					entryMap, _ := item.(map[string]any)
					if bash, ok := entryMap["bash"].(string); ok {
						entryMap["bash"] = strings.ReplaceAll(bash, ".github/hooks/", "hooks/")
					}
				}
				existing, _ := hooksMap[event].([]any)
				hooksMap[event] = append(existing, items...)
			}
		}
	}
	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(outDir, "hooks.json"), data, 0o644)
}

func copyDirIfExists(src, dst string) error {
	if !files.DirExists(src) {
		return nil
	}
	return files.CopyDir(src, dst)
}

func copyFileIfExists(src, dst string) error {
	if !files.FileExists(src) {
		return nil
	}
	return files.CopyFile(src, dst)
}

func copyLibraryFileIfExists(libFS fs.FS, src, dst string, mode os.FileMode) error {
	data, err := fs.ReadFile(libFS, src)
	if err != nil {
		if errorsIsNotExist(err) {
			return nil
		}
		return err
	}
	if err := files.EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}
	return os.WriteFile(dst, data, mode)
}

func errorsIsNotExist(err error) bool {
	return err == fs.ErrNotExist || os.IsNotExist(err)
}

func countFiles(root string) (int, error) {
	count := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}

func SupportedTargets() []string {
	out := []string{string(BundleTargetClaude), string(BundleTargetCopilotCLI), string(BundleTargetOmp), string(BundleTargetPi)}
	sort.Strings(out)
	return out
}

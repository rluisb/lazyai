package setupscan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/jsonc"
)

type recordedMCPEntry struct {
	Name        string `json:"name"`
	ConfigPath  string `json:"configPath"`
	Fingerprint string `json:"fingerprint"`
}

type observedMCPEntry struct {
	MCPEntry
	fingerprint string
}

func applyMCPEntryStates(entries *[]MCPEntry, recorded []recordedMCPEntry, observed []observedMCPEntry) []string {
	recordedByKey := make(map[string]recordedMCPEntry, len(recorded))
	for _, entry := range recorded {
		recordedByKey[mcpEntryKey(entry.ConfigPath, entry.Name)] = entry
	}
	observedByKey := make(map[string]observedMCPEntry, len(observed))
	for _, entry := range observed {
		observedByKey[mcpEntryKey(entry.ConfigPath, entry.Name)] = entry
	}

	var reasons []string
	for i := range *entries {
		entry := &(*entries)[i]
		recordedEntry, ok := recordedByKey[mcpEntryKey(entry.ConfigPath, entry.Name)]
		if !ok {
			entry.State = resourceStateConflict
			entry.Reasons = append(entry.Reasons, "unexpected-mcp-entry")
			reasons = append(reasons, fmt.Sprintf("unexpected-mcp-entry:%s:%s", entry.ConfigPath, entry.Name))
			continue
		}
		if observedByKey[mcpEntryKey(entry.ConfigPath, entry.Name)].fingerprint != recordedEntry.Fingerprint {
			entry.State = resourceStateConflict
			entry.Reasons = append(entry.Reasons, "mcp-entry-changed")
			reasons = append(reasons, fmt.Sprintf("mcp-entry-changed:%s:%s", entry.ConfigPath, entry.Name))
			continue
		}
		entry.State = resourceStateManaged
	}

	for _, recordedEntry := range recorded {
		if _, ok := observedByKey[mcpEntryKey(recordedEntry.ConfigPath, recordedEntry.Name)]; !ok {
			reasons = append(reasons, fmt.Sprintf("missing-mcp-entry:%s:%s", recordedEntry.ConfigPath, recordedEntry.Name))
		}
	}
	return reasons
}

func snapshotMCPEntries(rootPath string, observedFiles []string) ([]observedMCPEntry, error) {
	entries := make([]observedMCPEntry, 0)
	for _, relativePath := range observedFiles {
		fullPath := filepath.Join(rootPath, relativePath)
		if files.DirExists(fullPath) || !isMCPConfigCandidate(relativePath) {
			continue
		}
		parsed, err := readJSONLikeMap(fullPath)
		if err != nil {
			return nil, err
		}
		mcpMap := extractMCPEntriesMap(parsed, filepath.Base(relativePath))
		if len(mcpMap) == 0 {
			continue
		}
		names := make([]string, 0, len(mcpMap))
		for name := range mcpMap {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fingerprint, err := jsonFingerprint(mcpMap[name])
			if err != nil {
				return nil, err
			}
			entries = append(entries, observedMCPEntry{
				MCPEntry:    MCPEntry{Name: name, ConfigPath: filepath.ToSlash(relativePath)},
				fingerprint: fingerprint,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].ConfigPath != entries[j].ConfigPath {
			return entries[i].ConfigPath < entries[j].ConfigPath
		}
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}

func readJSONLikeMap(path string) (map[string]any, error) {
	if strings.HasSuffix(path, ".jsonc") {
		return jsonc.ReadJSONCFile(path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

func extractMCPEntriesMap(parsed map[string]any, baseName string) map[string]any {
	for _, key := range []string{"mcpServers", "mcp", "mcp_servers"} {
		if raw, ok := parsed[key].(map[string]any); ok {
			return raw
		}
	}
	if baseName == "mcp.json" {
		if raw, ok := parsed["servers"].(map[string]any); ok {
			return raw
		}
	}
	return nil
}

func jsonFingerprint(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func isMCPConfigCandidate(relativePath string) bool {
	base := filepath.Base(relativePath)
	switch base {
	case "settings.json", "settings.local.json", "mcp-config.json", "mcp.json", "opencode.json", "lazyai.mcp.jsonc":
		return true
	default:
		return false
	}
}

func toRecordedMCPEntries(entries []observedMCPEntry) []recordedMCPEntry {
	recorded := make([]recordedMCPEntry, 0, len(entries))
	for _, entry := range entries {
		recorded = append(recorded, recordedMCPEntry{Name: entry.Name, ConfigPath: entry.ConfigPath, Fingerprint: entry.fingerprint})
	}
	return recorded
}

func mcpEntryKey(configPath, name string) string {
	return configPath + "::" + name
}

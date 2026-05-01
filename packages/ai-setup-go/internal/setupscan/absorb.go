package setupscan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/jsonc"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

const (
	resourceStateManaged   = "managed"
	resourceStateAdoptable = "adoptable"
	resourceStateConflict  = "conflict"
	resourceStateUserOwned = "user-owned"
	resourceStateMissing   = "missing"

	scanRegistryVersion = 1
)

var reservedContextMarkdownFiles = map[string]bool{
	"AGENTS.md":               true,
	"CLAUDE.md":               true,
	"copilot-instructions.md": true,
}

type OperationResult struct {
	Mode         string             `json:"mode"`
	RegistryPath string             `json:"registryPath"`
	ImportRoot   string             `json:"importRoot,omitempty"`
	Backups      []string           `json:"backups,omitempty"`
	Adopted      []OperationTarget  `json:"adopted,omitempty"`
	Imported     []ImportedResource `json:"imported,omitempty"`
	Skipped      []SkippedOperation `json:"skipped,omitempty"`
}

type OperationTarget struct {
	TargetID string `json:"targetId"`
	Scope    string `json:"scope"`
	Origin   string `json:"origin"`
	RootPath string `json:"rootPath"`
}

type ImportedResource struct {
	TargetID        string `json:"targetId"`
	Scope           string `json:"scope"`
	Origin          string `json:"origin"`
	SourcePath      string `json:"sourcePath"`
	DestinationPath string `json:"destinationPath"`
}

type SkippedOperation struct {
	TargetID string `json:"targetId"`
	Scope    string `json:"scope"`
	Origin   string `json:"origin"`
	RootPath string `json:"rootPath"`
	State    string `json:"state"`
	Reason   string `json:"reason"`
}

type scanRegistry struct {
	Version   int               `json:"version"`
	Resources []managedResource `json:"resources,omitempty"`
	Imports   []importRecord    `json:"imports,omitempty"`
}

type managedResource struct {
	TargetID      string             `json:"targetId"`
	Scope         string             `json:"scope"`
	Origin        string             `json:"origin"`
	RootPath      string             `json:"rootPath"`
	State         string             `json:"state"`
	ObservedPaths []recordedPath     `json:"observedPaths,omitempty"`
	MCPEntries    []recordedMCPEntry `json:"mcpEntries,omitempty"`
	UpdatedAt     string             `json:"updatedAt"`
}

type importRecord struct {
	TargetID        string   `json:"targetId"`
	Scope           string   `json:"scope"`
	Origin          string   `json:"origin"`
	RootPath        string   `json:"rootPath"`
	ImportedPaths   []string `json:"importedPaths,omitempty"`
	DestinationRoot string   `json:"destinationRoot"`
	UpdatedAt       string   `json:"updatedAt"`
}

type recordedPath struct {
	RelativePath string `json:"relativePath"`
	Fingerprint  string `json:"fingerprint"`
}

type recordedMCPEntry struct {
	Name        string `json:"name"`
	ConfigPath  string `json:"configPath"`
	Fingerprint string `json:"fingerprint"`
}

type observedMCPEntry struct {
	MCPEntry
	fingerprint string
}

func Run(opts Options) (*Inventory, error) {
	inventory, err := Scan(opts)
	if err != nil {
		return nil, err
	}
	if !opts.Adopt && !opts.Import {
		return inventory, nil
	}

	registryPath := scanRegistryPath(aiSetupHome(opts))
	registry, err := loadRegistry(aiSetupHome(opts))
	if err != nil {
		return nil, err
	}

	operation := &OperationResult{
		Mode:         operationMode(opts),
		RegistryPath: registryPath,
		ImportRoot:   importsRoot(aiSetupHome(opts)),
	}
	backupSet := map[string]bool{}

	if opts.Adopt {
		adoptResources(inventory, registry, operation)
	}
	if opts.Import {
		if err := importResources(inventory, registry, operation, backupSet); err != nil {
			return nil, err
		}
	}

	if err := writeRegistryIfChanged(registryPath, registry, operation, backupSet); err != nil {
		return nil, err
	}

	rescan, err := Scan(Options{HomeDir: opts.HomeDir, TargetDir: opts.TargetDir})
	if err != nil {
		return nil, err
	}
	rescan.Operation = operation
	return rescan, nil
}

func applyDetectionState(detection *TargetDetection, tool types.ToolId, registry *scanRegistry) {
	if detection.Status == "missing" {
		detection.State = resourceStateMissing
		return
	}

	observedPaths, err := snapshotObservedPaths(detection.RootPath, detection.ObservedFiles)
	if err != nil {
		detection.State = resourceStateConflict
		detection.Reasons = append(detection.Reasons, fmt.Sprintf("snapshot-failed:%v", err))
		return
	}
	mcpEntries, err := snapshotMCPEntries(detection.RootPath, detection.ObservedFiles)
	if err != nil {
		detection.State = resourceStateConflict
		detection.Reasons = append(detection.Reasons, fmt.Sprintf("mcp-parse-failed:%v", err))
		return
	}
	for _, entry := range mcpEntries {
		detection.MCPEntries = append(detection.MCPEntries, entry.MCPEntry)
	}

	record := registry.findResource(string(tool), detection.Scope, detection.Origin, detection.RootPath)
	if record == nil {
		detection.State = resourceStateAdoptable
		for i := range detection.MCPEntries {
			detection.MCPEntries[i].State = resourceStateAdoptable
		}
		return
	}

	switch record.State {
	case resourceStateUserOwned:
		detection.State = resourceStateUserOwned
		for i := range detection.MCPEntries {
			detection.MCPEntries[i].State = resourceStateUserOwned
		}
		return
	case resourceStateManaged:
		pathConflictReasons := compareObservedPaths(record.ObservedPaths, observedPaths)
		entryConflictReasons := applyMCPEntryStates(&detection.MCPEntries, record.MCPEntries, mcpEntries)
		detection.Reasons = append(detection.Reasons, pathConflictReasons...)
		detection.Reasons = append(detection.Reasons, entryConflictReasons...)
		if len(detection.Reasons) == 0 {
			detection.State = resourceStateManaged
			return
		}
		detection.State = resourceStateConflict
		return
	default:
		detection.State = resourceStateConflict
		detection.Reasons = append(detection.Reasons, "unsupported-registry-state")
		for i := range detection.MCPEntries {
			detection.MCPEntries[i].State = resourceStateConflict
		}
	}
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

func compareObservedPaths(recorded []recordedPath, observed []recordedPath) []string {
	recordedByPath := make(map[string]string, len(recorded))
	for _, path := range recorded {
		recordedByPath[path.RelativePath] = path.Fingerprint
	}
	observedByPath := make(map[string]string, len(observed))
	for _, path := range observed {
		observedByPath[path.RelativePath] = path.Fingerprint
	}

	var reasons []string
	for relativePath, fingerprint := range recordedByPath {
		observedFingerprint, ok := observedByPath[relativePath]
		if !ok {
			reasons = append(reasons, "missing-path:"+relativePath)
			continue
		}
		if observedFingerprint != fingerprint {
			reasons = append(reasons, "changed-path:"+relativePath)
		}
	}
	for relativePath := range observedByPath {
		if _, ok := recordedByPath[relativePath]; !ok {
			reasons = append(reasons, "unexpected-path:"+relativePath)
		}
	}
	sort.Strings(reasons)
	return reasons
}

func adoptResources(inventory *Inventory, registry *scanRegistry, operation *OperationResult) {
	now := time.Now().UTC().Format(time.RFC3339)
	seenRoots := map[string]bool{}
	for _, target := range inventory.CurrentState.Targets {
		for _, detection := range target.Detections {
			rootKey := target.ID + "::" + detection.RootPath
			if seenRoots[rootKey] {
				continue
			}
			if detection.State != resourceStateAdoptable {
				operation.Skipped = append(operation.Skipped, SkippedOperation{
					TargetID: target.ID,
					Scope:    detection.Scope,
					Origin:   detection.Origin,
					RootPath: detection.RootPath,
					State:    detection.State,
					Reason:   "not-adoptable",
				})
				continue
			}

			observedPaths, err := snapshotObservedPaths(detection.RootPath, detection.ObservedFiles)
			if err != nil {
				operation.Skipped = append(operation.Skipped, SkippedOperation{TargetID: target.ID, Scope: detection.Scope, Origin: detection.Origin, RootPath: detection.RootPath, State: detection.State, Reason: err.Error()})
				continue
			}
			mcpEntries, err := snapshotMCPEntries(detection.RootPath, detection.ObservedFiles)
			if err != nil {
				operation.Skipped = append(operation.Skipped, SkippedOperation{TargetID: target.ID, Scope: detection.Scope, Origin: detection.Origin, RootPath: detection.RootPath, State: detection.State, Reason: err.Error()})
				continue
			}
			record := managedResource{
				TargetID:      target.ID,
				Scope:         detection.Scope,
				Origin:        detection.Origin,
				RootPath:      detection.RootPath,
				State:         resourceStateManaged,
				ObservedPaths: observedPaths,
				MCPEntries:    toRecordedMCPEntries(mcpEntries),
				UpdatedAt:     now,
			}
			registry.upsertResource(record)
			operation.Adopted = append(operation.Adopted, OperationTarget{TargetID: target.ID, Scope: detection.Scope, Origin: detection.Origin, RootPath: detection.RootPath})
			seenRoots[rootKey] = true
		}
	}
}

func importResources(inventory *Inventory, registry *scanRegistry, operation *OperationResult, backupSet map[string]bool) error {
	now := time.Now().UTC().Format(time.RFC3339)
	importRoot := operation.ImportRoot
	if err := files.EnsureDir(importRoot); err != nil {
		return err
	}
	seenRoots := map[string]bool{}

	for _, target := range inventory.CurrentState.Targets {
		for _, detection := range target.Detections {
			rootKey := target.ID + "::" + detection.RootPath
			if seenRoots[rootKey] {
				continue
			}
			if !canImportState(detection.State) {
				operation.Skipped = append(operation.Skipped, SkippedOperation{
					TargetID: target.ID,
					Scope:    detection.Scope,
					Origin:   detection.Origin,
					RootPath: detection.RootPath,
					State:    detection.State,
					Reason:   "not-importable",
				})
				continue
			}

			destinationRoot := filepath.Join(importRoot, target.ID, importDirectoryName(detection.Scope, detection.Origin, detection.RootPath))
			importedPaths := make([]string, 0, len(detection.ObservedFiles))
			for _, relativePath := range detection.ObservedFiles {
				sourcePath := filepath.Join(detection.RootPath, relativePath)
				destinationPath := filepath.Join(destinationRoot, relativePath)
				if err := importPath(sourcePath, destinationPath, operation, backupSet); err != nil {
					return err
				}
				operation.Imported = append(operation.Imported, ImportedResource{
					TargetID:        target.ID,
					Scope:           detection.Scope,
					Origin:          detection.Origin,
					SourcePath:      sourcePath,
					DestinationPath: destinationPath,
				})
				importedPaths = append(importedPaths, filepath.ToSlash(relativePath))
			}
			registry.upsertImport(importRecord{
				TargetID:        target.ID,
				Scope:           detection.Scope,
				Origin:          detection.Origin,
				RootPath:        detection.RootPath,
				ImportedPaths:   importedPaths,
				DestinationRoot: destinationRoot,
				UpdatedAt:       now,
			})
			seenRoots[rootKey] = true
		}
	}
	return nil
}

func importPath(sourcePath, destinationPath string, operation *OperationResult, backupSet map[string]bool) error {
	if !files.FileExists(sourcePath) {
		return nil
	}
	if same, err := pathsMatch(sourcePath, destinationPath); err == nil && same {
		return nil
	}
	if files.FileExists(destinationPath) {
		if err := backupIfNeeded(destinationPath, operation, backupSet); err != nil {
			return err
		}
		if err := files.RemoveAll(destinationPath); err != nil {
			return err
		}
	}
	return copyPath(sourcePath, destinationPath)
}

func backupIfNeeded(path string, operation *OperationResult, backupSet map[string]bool) error {
	if backupSet[path] {
		return nil
	}
	backupPath, err := files.CreateTimestampedBackup(path)
	if err != nil {
		return err
	}
	backupSet[path] = true
	operation.Backups = append(operation.Backups, backupPath)
	return nil
}

func copyPath(sourcePath, destinationPath string) error {
	if files.DirExists(sourcePath) {
		return copyDirSkippingReservedContextDocs(sourcePath, destinationPath)
	}
	return files.CopyFile(sourcePath, destinationPath)
}

func copyDirSkippingReservedContextDocs(sourcePath, destinationPath string) error {
	return filepath.WalkDir(sourcePath, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(sourcePath, current)
		if err != nil {
			return err
		}
		if rel == "." {
			return files.EnsureDir(destinationPath)
		}
		if !entry.IsDir() && reservedContextMarkdownFiles[entry.Name()] {
			return nil
		}

		dest := filepath.Join(destinationPath, rel)
		if entry.IsDir() {
			return files.EnsureDir(dest)
		}
		return files.CopyFile(current, dest)
	})
}

func pathsMatch(sourcePath, destinationPath string) (bool, error) {
	if !files.FileExists(destinationPath) {
		return false, nil
	}
	sourceFingerprint, err := pathFingerprint(sourcePath)
	if err != nil {
		return false, err
	}
	destinationFingerprint, err := pathFingerprint(destinationPath)
	if err != nil {
		return false, err
	}
	return sourceFingerprint == destinationFingerprint, nil
}

func pathFingerprint(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return files.FileHash(path)
	}

	entries := []string{}
	err = filepath.WalkDir(path, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == path {
			return nil
		}
		if !entry.IsDir() && reservedContextMarkdownFiles[entry.Name()] {
			return nil
		}
		rel, err := filepath.Rel(path, current)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if entry.IsDir() {
			entries = append(entries, "dir:"+rel)
			return nil
		}
		hash, err := files.FileHash(current)
		if err != nil {
			return err
		}
		entries = append(entries, "file:"+rel+":"+hash)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(entries)
	sum := sha256.Sum256([]byte(strings.Join(entries, "\n")))
	return hex.EncodeToString(sum[:]), nil
}

func snapshotObservedPaths(rootPath string, observedFiles []string) ([]recordedPath, error) {
	paths := make([]recordedPath, 0, len(observedFiles))
	for _, relativePath := range observedFiles {
		fullPath := filepath.Join(rootPath, relativePath)
		fingerprint, err := pathFingerprint(fullPath)
		if err != nil {
			return nil, err
		}
		paths = append(paths, recordedPath{RelativePath: filepath.ToSlash(relativePath), Fingerprint: fingerprint})
	}
	sort.Slice(paths, func(i, j int) bool { return paths[i].RelativePath < paths[j].RelativePath })
	return paths, nil
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
	case "settings.json", "settings.local.json", "mcp-config.json", "mcp.json", "opencode.json", "opencode.jsonc":
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

func operationMode(opts Options) string {
	switch {
	case opts.Adopt && opts.Import:
		return "adopt-import"
	case opts.Adopt:
		return "adopt"
	case opts.Import:
		return "import"
	default:
		return "scan"
	}
}

func aiSetupHome(opts Options) string {
	return filepath.Join(opts.HomeDir, ".ai-setup")
}

func scanRegistryPath(aiSetupHome string) string {
	return filepath.Join(aiSetupHome, "config", "setup-scan-registry.json")
}

func importsRoot(aiSetupHome string) string {
	return filepath.Join(aiSetupHome, "imports")
}

func importDirectoryName(scope, origin, rootPath string) string {
	hash := sha256.Sum256([]byte(rootPath))
	return fmt.Sprintf("%s-%s-%s", scope, origin, hex.EncodeToString(hash[:])[:12])
}

func loadRegistry(aiSetupHome string) (*scanRegistry, error) {
	path := scanRegistryPath(aiSetupHome)
	if !files.FileExists(path) {
		return &scanRegistry{Version: scanRegistryVersion}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read setup scan registry: %w", err)
	}
	var registry scanRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("parse setup scan registry: %w", err)
	}
	if registry.Version == 0 {
		registry.Version = scanRegistryVersion
	}
	registry.sort()
	return &registry, nil
}

func writeRegistryIfChanged(path string, registry *scanRegistry, operation *OperationResult, backupSet map[string]bool) error {
	registry.Version = scanRegistryVersion
	registry.sort()
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal setup scan registry: %w", err)
	}
	data = append(data, '\n')

	if existing, err := os.ReadFile(path); err == nil && string(existing) == string(data) {
		sortOperation(operation)
		return nil
	}
	if err := files.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	if files.FileExists(path) {
		if err := backupIfNeeded(path, operation, backupSet); err != nil {
			return err
		}
	}
	if err := files.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	sortOperation(operation)
	return nil
}

func sortOperation(operation *OperationResult) {
	sort.Strings(operation.Backups)
	sort.Slice(operation.Adopted, func(i, j int) bool { return operationTargetLess(operation.Adopted[i], operation.Adopted[j]) })
	sort.Slice(operation.Imported, func(i, j int) bool {
		if operation.Imported[i].TargetID != operation.Imported[j].TargetID {
			return operation.Imported[i].TargetID < operation.Imported[j].TargetID
		}
		if operation.Imported[i].Scope != operation.Imported[j].Scope {
			return operation.Imported[i].Scope < operation.Imported[j].Scope
		}
		if operation.Imported[i].Origin != operation.Imported[j].Origin {
			return operation.Imported[i].Origin < operation.Imported[j].Origin
		}
		if operation.Imported[i].SourcePath != operation.Imported[j].SourcePath {
			return operation.Imported[i].SourcePath < operation.Imported[j].SourcePath
		}
		return operation.Imported[i].DestinationPath < operation.Imported[j].DestinationPath
	})
	sort.Slice(operation.Skipped, func(i, j int) bool {
		if operation.Skipped[i].TargetID != operation.Skipped[j].TargetID {
			return operation.Skipped[i].TargetID < operation.Skipped[j].TargetID
		}
		if operation.Skipped[i].Scope != operation.Skipped[j].Scope {
			return operation.Skipped[i].Scope < operation.Skipped[j].Scope
		}
		if operation.Skipped[i].Origin != operation.Skipped[j].Origin {
			return operation.Skipped[i].Origin < operation.Skipped[j].Origin
		}
		if operation.Skipped[i].RootPath != operation.Skipped[j].RootPath {
			return operation.Skipped[i].RootPath < operation.Skipped[j].RootPath
		}
		return operation.Skipped[i].Reason < operation.Skipped[j].Reason
	})
}

func operationTargetLess(left, right OperationTarget) bool {
	if left.TargetID != right.TargetID {
		return left.TargetID < right.TargetID
	}
	if left.Scope != right.Scope {
		return left.Scope < right.Scope
	}
	if left.Origin != right.Origin {
		return left.Origin < right.Origin
	}
	return left.RootPath < right.RootPath
}

func canImportState(state string) bool {
	switch state {
	case resourceStateAdoptable, resourceStateConflict, resourceStateUserOwned:
		return true
	default:
		return false
	}
}

func mcpEntryKey(configPath, name string) string {
	return configPath + "::" + name
}

func (registry *scanRegistry) findResource(targetID, scope, origin, rootPath string) *managedResource {
	for i := range registry.Resources {
		record := &registry.Resources[i]
		if record.TargetID == targetID && record.Scope == scope && record.Origin == origin && record.RootPath == rootPath {
			return record
		}
	}
	for i := range registry.Resources {
		record := &registry.Resources[i]
		if record.TargetID == targetID && record.RootPath == rootPath {
			return record
		}
	}
	return nil
}

func (registry *scanRegistry) upsertResource(record managedResource) {
	for i := range registry.Resources {
		if registry.Resources[i].TargetID == record.TargetID && registry.Resources[i].Scope == record.Scope && registry.Resources[i].Origin == record.Origin && registry.Resources[i].RootPath == record.RootPath {
			registry.Resources[i] = record
			return
		}
	}
	registry.Resources = append(registry.Resources, record)
}

func (registry *scanRegistry) upsertImport(record importRecord) {
	for i := range registry.Imports {
		if registry.Imports[i].TargetID == record.TargetID && registry.Imports[i].Scope == record.Scope && registry.Imports[i].Origin == record.Origin && registry.Imports[i].RootPath == record.RootPath {
			registry.Imports[i] = record
			return
		}
	}
	registry.Imports = append(registry.Imports, record)
}

func (registry *scanRegistry) sort() {
	sort.Slice(registry.Resources, func(i, j int) bool {
		left, right := registry.Resources[i], registry.Resources[j]
		if left.TargetID != right.TargetID {
			return left.TargetID < right.TargetID
		}
		if left.Scope != right.Scope {
			return left.Scope < right.Scope
		}
		if left.Origin != right.Origin {
			return left.Origin < right.Origin
		}
		return left.RootPath < right.RootPath
	})
	for i := range registry.Resources {
		sort.Slice(registry.Resources[i].ObservedPaths, func(a, b int) bool {
			return registry.Resources[i].ObservedPaths[a].RelativePath < registry.Resources[i].ObservedPaths[b].RelativePath
		})
		sort.Slice(registry.Resources[i].MCPEntries, func(a, b int) bool {
			if registry.Resources[i].MCPEntries[a].ConfigPath != registry.Resources[i].MCPEntries[b].ConfigPath {
				return registry.Resources[i].MCPEntries[a].ConfigPath < registry.Resources[i].MCPEntries[b].ConfigPath
			}
			return registry.Resources[i].MCPEntries[a].Name < registry.Resources[i].MCPEntries[b].Name
		})
	}

	sort.Slice(registry.Imports, func(i, j int) bool {
		left, right := registry.Imports[i], registry.Imports[j]
		if left.TargetID != right.TargetID {
			return left.TargetID < right.TargetID
		}
		if left.Scope != right.Scope {
			return left.Scope < right.Scope
		}
		if left.Origin != right.Origin {
			return left.Origin < right.Origin
		}
		return left.RootPath < right.RootPath
	})
	for i := range registry.Imports {
		sort.Strings(registry.Imports[i].ImportedPaths)
	}
}

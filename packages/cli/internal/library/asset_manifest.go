package library

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	LibraryRootRelPath        = "packages/cli/library"
	CanonicalRootRelPath      = "packages/cli/library/canonical"
	ProvenanceManifestRelPath = "packages/cli/library/manifests/provenance.yaml"
	CurationManifestRelPath   = "packages/cli/library/manifests/curation.yaml"
)

// CurationCoverageRoots are the active embedded asset families guarded by the
// curation manifest. Command/chatmode/prompt inventories are intentionally not
// mixed into this Wave 0 check.
var CurationCoverageRoots = []string{
	"canonical/agents",
	"canonical/hooks",
	"canonical/skills",
	"hooks",
	"root",
	"rules",
	"skills",
	"standards",
	"templates",
	"tool-agents",
	"tool-templates",
}

// ProvenanceManifest records where canonical embedded assets came from and the
// local hash that makes silent drift visible.
type ProvenanceManifest struct {
	Version       int                   `yaml:"version"`
	CanonicalRoot string                `yaml:"canonical_root"`
	Entries       []ProvenanceEntry     `yaml:"entries"`
	Exclusions    []ProvenanceExclusion `yaml:"exclusions"`
}

// ProvenanceEntry covers one active file under packages/cli/library/canonical.
type ProvenanceEntry struct {
	Path        string `yaml:"path"`
	Kind        string `yaml:"kind"`
	SourceRepo  string `yaml:"source_repo"`
	SourceRef   string `yaml:"source_ref"`
	SourcePath  string `yaml:"source_path"`
	LocalSHA256 string `yaml:"local_sha256"`
	Mode        string `yaml:"mode"`
	Notes       string `yaml:"notes"`
}

// ProvenanceExclusion is an explicit decision not to hash-cover a canonical file.
type ProvenanceExclusion struct {
	Path   string `yaml:"path"`
	Kind   string `yaml:"kind"`
	Reason string `yaml:"reason"`
}

// CurationManifest records why LazyAI embeds each guarded asset and where it is emitted.
type CurationManifest struct {
	Version     int                 `yaml:"version"`
	LibraryRoot string              `yaml:"library_root"`
	Entries     []CurationEntry     `yaml:"entries"`
	Exclusions  []CurationExclusion `yaml:"exclusions"`
}

// CurationEntry covers one active embedded asset in a guarded family.
type CurationEntry struct {
	Path                      string   `yaml:"path"`
	Kind                      string   `yaml:"kind"`
	Category                  string   `yaml:"category"`
	AdapterTargets            []string `yaml:"adapter_targets"`
	ReasonKept                string   `yaml:"reason_kept"`
	ReasonCompressedOrChanged string   `yaml:"reason_compressed_or_changed"`
	ReasonDropped             string   `yaml:"reason_dropped"`
	TokenRentRelevant         *bool    `yaml:"token_rent_relevant"`
}

// CurationExclusion records an upstream concept or local asset excluded from curation.
type CurationExclusion struct {
	Path          string `yaml:"path"`
	Kind          string `yaml:"kind"`
	Category      string `yaml:"category"`
	ReasonDropped string `yaml:"reason_dropped"`
}

// ManifestValidationError reports every manifest problem in one failure.
type ManifestValidationError struct {
	Problems []string
}

func (e *ManifestValidationError) Error() string {
	return "library manifest validation failed:\n- " + strings.Join(e.Problems, "\n- ")
}

// ValidateProjectAssetManifests validates repository-local asset manifests.
func ValidateProjectAssetManifests(projectRoot string) error {
	var problems []string
	if err := ValidateProvenanceManifest(projectRoot); err != nil {
		appendManifestProblems(&problems, err)
	}
	if err := ValidateCurationManifest(projectRoot); err != nil {
		appendManifestProblems(&problems, err)
	}
	if len(problems) > 0 {
		return &ManifestValidationError{Problems: problems}
	}
	return nil
}

// ValidateProvenanceManifest checks canonical coverage and local hashes using
// only files inside projectRoot.
func ValidateProvenanceManifest(projectRoot string) error {
	manifest, err := ReadProvenanceManifest(filepath.Join(projectRoot, ProvenanceManifestRelPath))
	if err != nil {
		return err
	}

	var problems []string
	if manifest.Version != 1 {
		problems = append(problems, fmt.Sprintf("%s: version=%d, want 1", ProvenanceManifestRelPath, manifest.Version))
	}
	if manifest.CanonicalRoot != CanonicalRootRelPath {
		problems = append(problems, fmt.Sprintf("%s: canonical_root=%q, want %q", ProvenanceManifestRelPath, manifest.CanonicalRoot, CanonicalRootRelPath))
	}

	covered := map[string]string{}
	for i, entry := range manifest.Entries {
		prefix := fmt.Sprintf("%s: entries[%d]", ProvenanceManifestRelPath, i)
		clean := validateManifestEntryPath(prefix, entry.Path, CanonicalRootRelPath, &problems)
		requireText(&problems, prefix, "kind", entry.Kind)
		requireText(&problems, prefix, "source_repo", entry.SourceRepo)
		requireText(&problems, prefix, "source_ref", entry.SourceRef)
		requireText(&problems, prefix, "source_path", entry.SourcePath)
		requireText(&problems, prefix, "local_sha256", entry.LocalSHA256)
		requireText(&problems, prefix, "mode", entry.Mode)
		requireText(&problems, prefix, "notes", entry.Notes)
		if clean == "" {
			continue
		}
		markCovered(covered, clean, prefix, &problems)
		if len(entry.LocalSHA256) != 64 {
			problems = append(problems, fmt.Sprintf("%s: local_sha256 must be 64 hex characters", prefix))
		}
		actual, err := fileSHA256(filepath.Join(projectRoot, clean))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: hash file: %v", prefix, err))
			continue
		}
		if actual != entry.LocalSHA256 {
			problems = append(problems, fmt.Sprintf("%s: stale local_sha256 for %s: got %s, want %s", prefix, clean, entry.LocalSHA256, actual))
		}
	}
	for i, exclusion := range manifest.Exclusions {
		prefix := fmt.Sprintf("%s: exclusions[%d]", ProvenanceManifestRelPath, i)
		clean := validateManifestEntryPath(prefix, exclusion.Path, CanonicalRootRelPath, &problems)
		requireText(&problems, prefix, "kind", exclusion.Kind)
		requireText(&problems, prefix, "reason", exclusion.Reason)
		if clean != "" {
			markCovered(covered, clean, prefix, &problems)
		}
	}

	files, err := walkRepoFiles(projectRoot, CanonicalRootRelPath)
	if err != nil {
		problems = append(problems, err.Error())
	}
	for _, file := range files {
		if _, ok := covered[file]; !ok {
			problems = append(problems, fmt.Sprintf("missing provenance manifest coverage: %s", file))
		}
	}

	if len(problems) > 0 {
		return &ManifestValidationError{Problems: problems}
	}
	return nil
}

// ValidateCurationManifest checks guarded embedded asset families using only
// files inside projectRoot. It does not read vibe-lab or token-rent state.
func ValidateCurationManifest(projectRoot string) error {
	manifest, err := ReadCurationManifest(filepath.Join(projectRoot, CurationManifestRelPath))
	if err != nil {
		return err
	}

	var problems []string
	if manifest.Version != 1 {
		problems = append(problems, fmt.Sprintf("%s: version=%d, want 1", CurationManifestRelPath, manifest.Version))
	}
	if manifest.LibraryRoot != LibraryRootRelPath {
		problems = append(problems, fmt.Sprintf("%s: library_root=%q, want %q", CurationManifestRelPath, manifest.LibraryRoot, LibraryRootRelPath))
	}

	covered := map[string]string{}
	for i, entry := range manifest.Entries {
		prefix := fmt.Sprintf("%s: entries[%d]", CurationManifestRelPath, i)
		clean := validateManifestEntryPath(prefix, entry.Path, LibraryRootRelPath, &problems)
		requireText(&problems, prefix, "kind", entry.Kind)
		requireText(&problems, prefix, "category", entry.Category)
		requireText(&problems, prefix, "reason_kept", entry.ReasonKept)
		requireText(&problems, prefix, "reason_compressed_or_changed", entry.ReasonCompressedOrChanged)
		if entry.TokenRentRelevant == nil {
			problems = append(problems, fmt.Sprintf("%s: token_rent_relevant is required", prefix))
		}
		if len(entry.AdapterTargets) == 0 {
			problems = append(problems, fmt.Sprintf("%s: adapter_targets must name tools or none", prefix))
		}
		for j, target := range entry.AdapterTargets {
			if strings.TrimSpace(target) == "" {
				problems = append(problems, fmt.Sprintf("%s: adapter_targets[%d] is empty", prefix, j))
			}
		}
		if clean == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(projectRoot, clean)); err != nil {
			problems = append(problems, fmt.Sprintf("%s: entry path is not an active repository file: %v", prefix, err))
		}
		markCovered(covered, clean, prefix, &problems)
	}
	for i, exclusion := range manifest.Exclusions {
		prefix := fmt.Sprintf("%s: exclusions[%d]", CurationManifestRelPath, i)
		clean := validateManifestEntryPath(prefix, exclusion.Path, LibraryRootRelPath, &problems)
		requireText(&problems, prefix, "kind", exclusion.Kind)
		requireText(&problems, prefix, "category", exclusion.Category)
		requireText(&problems, prefix, "reason_dropped", exclusion.ReasonDropped)
		if clean != "" {
			markCovered(covered, clean, prefix, &problems)
		}
	}

	for _, root := range CurationCoverageRoots {
		files, err := walkRepoFilesIfExists(projectRoot, filepath.ToSlash(filepath.Join(LibraryRootRelPath, root)))
		if err != nil {
			problems = append(problems, err.Error())
		}
		for _, file := range files {
			if _, ok := covered[file]; !ok {
				problems = append(problems, fmt.Sprintf("missing curation manifest coverage: %s", file))
			}
		}
	}

	if len(problems) > 0 {
		return &ManifestValidationError{Problems: problems}
	}
	return nil
}

// ReadProvenanceManifest parses a provenance manifest from disk.
func ReadProvenanceManifest(path string) (ProvenanceManifest, error) {
	var manifest ProvenanceManifest
	if err := readYAML(path, &manifest); err != nil {
		return ProvenanceManifest{}, err
	}
	return manifest, nil
}

// ReadCurationManifest parses a curation manifest from disk.
func ReadCurationManifest(path string) (CurationManifest, error) {
	var manifest CurationManifest
	if err := readYAML(path, &manifest); err != nil {
		return CurationManifest{}, err
	}
	return manifest, nil
}

func readYAML(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read library manifest %s: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("parse library manifest %s: %w", path, err)
	}
	return nil
}

func walkRepoFiles(projectRoot, rootRel string) ([]string, error) {
	return walkRepoFilesUnder(projectRoot, rootRel, false)
}

func walkRepoFilesIfExists(projectRoot, rootRel string) ([]string, error) {
	return walkRepoFilesUnder(projectRoot, rootRel, true)
}

func walkRepoFilesUnder(projectRoot, rootRel string, skipMissing bool) ([]string, error) {
	root := filepath.Join(projectRoot, filepath.FromSlash(rootRel))
	if skipMissing {
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, fmt.Errorf("stat curation root %s: %w", rootRel, err)
		}
	}
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}
		repoRel := filepath.ToSlash(rel)
		if excludedLibraryManifestFile(repoRel) {
			return nil
		}
		files = append(files, repoRel)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk library manifest root %s: %w", rootRel, err)
	}
	sort.Strings(files)
	return files, nil
}

func excludedLibraryManifestFile(repoRel string) bool {
	for _, part := range strings.Split(repoRel, "/") {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

func validateManifestEntryPath(prefix, path, root string, problems *[]string) string {
	if strings.TrimSpace(path) == "" {
		*problems = append(*problems, fmt.Sprintf("%s: path is required", prefix))
		return ""
	}
	if filepath.IsAbs(path) {
		*problems = append(*problems, fmt.Sprintf("%s: path must be repository-relative, got %q", prefix, path))
		return ""
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		*problems = append(*problems, fmt.Sprintf("%s: path escapes repository root: %q", prefix, path))
		return ""
	}
	if clean != root && !strings.HasPrefix(clean, root+"/") {
		*problems = append(*problems, fmt.Sprintf("%s: path %q must be under %s", prefix, clean, root))
		return ""
	}
	return clean
}

func requireText(problems *[]string, prefix, field, value string) {
	if strings.TrimSpace(value) == "" {
		*problems = append(*problems, fmt.Sprintf("%s: %s is required", prefix, field))
	}
}

func markCovered(covered map[string]string, path, prefix string, problems *[]string) {
	if previous, ok := covered[path]; ok {
		*problems = append(*problems, fmt.Sprintf("%s: duplicate coverage for %s; first covered by %s", prefix, path, previous))
		return
	}
	covered[path] = prefix
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func appendManifestProblems(problems *[]string, err error) {
	if validation, ok := err.(*ManifestValidationError); ok {
		*problems = append(*problems, validation.Problems...)
		return
	}
	*problems = append(*problems, err.Error())
}

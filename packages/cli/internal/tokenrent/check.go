package tokenrent

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultBudgetBytes = 50000
	CanonicalSubdir    = "packages/cli/library/canonical"
	OverrideRelPath    = ".lazyai/token-rent-override"
)

// Result captures a canonical-library budget evaluation.
type Result struct {
	CanonicalDir string
	OverridePath string
	TotalBytes   int
	BudgetBytes  int
	Files        []CountedFile
	Override     *Override
	OverrideUsed bool
}

// CountedFile is one counted canonical asset.
type CountedFile struct {
	Path  string
	Bytes int
}

// Override is the documented escape hatch for exceeding the default budget.
type Override struct {
	Budget     int    `yaml:"budget"`
	Reason     string `yaml:"reason"`
	ApprovedBy string `yaml:"approved_by"`
	Expires    string `yaml:"expires"`
}

// BudgetError reports a default-budget breach without a valid override.
type BudgetError struct {
	TotalBytes  int
	BudgetBytes int
}

func (e *BudgetError) Error() string {
	return fmt.Sprintf("Library budget exceeded: %d / %d bytes. Override: add .lazyai/token-rent-override with justification.", e.TotalBytes, e.BudgetBytes)
}

// OverrideError reports an invalid override file.
type OverrideError struct {
	Path   string
	Reason string
}

func (e *OverrideError) Error() string {
	if e.Path == "" {
		return e.Reason
	}
	return fmt.Sprintf("invalid token-rent override %s: %s", e.Path, e.Reason)
}

// CheckProject evaluates the canonical library budget for a repository root.
func CheckProject(projectRoot string) (Result, error) {
	canonicalDir := filepath.Join(projectRoot, CanonicalSubdir)
	overridePath := filepath.Join(projectRoot, OverrideRelPath)
	return Check(canonicalDir, overridePath, DefaultBudgetBytes)
}

// Check evaluates a canonical library path against a budget and optional override.
func Check(canonicalDir, overridePath string, budgetBytes int) (Result, error) {
	result := Result{
		CanonicalDir: canonicalDir,
		OverridePath: overridePath,
		BudgetBytes:  budgetBytes,
	}

	files, totalBytes, err := countCanonicalFiles(canonicalDir)
	if err != nil {
		return result, err
	}
	result.Files = files
	result.TotalBytes = totalBytes

	if totalBytes <= budgetBytes {
		return result, nil
	}

	override, err := readOverride(overridePath)
	if err != nil {
		var invalid *OverrideError
		if errors.As(err, &invalid) {
			return result, err
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return result, err
		}
		return result, &BudgetError{TotalBytes: totalBytes, BudgetBytes: budgetBytes}
	}

	result.Override = override
	result.OverrideUsed = true
	if override.Budget > 0 && totalBytes > override.Budget {
		return result, &BudgetError{TotalBytes: totalBytes, BudgetBytes: override.Budget}
	}
	return result, nil
}

func countCanonicalFiles(canonicalDir string) ([]CountedFile, int, error) {
	var files []CountedFile
	total := 0

	info, err := os.Stat(canonicalDir)
	if err != nil {
		return nil, 0, fmt.Errorf("stat canonical library: %w", err)
	}
	if !info.IsDir() {
		return nil, 0, fmt.Errorf("canonical library path is not a directory: %s", canonicalDir)
	}

	err = filepath.WalkDir(canonicalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if excludedFromBudget(path) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(canonicalDir, path)
		if err != nil {
			return err
		}
		files = append(files, CountedFile{Path: filepath.ToSlash(rel), Bytes: int(info.Size())})
		total += int(info.Size())
		return nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("walk canonical library: %w", err)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, total, nil
}

func excludedFromBudget(path string) bool {
	base := filepath.Base(path)
	if base == ".gitkeep" {
		return true
	}
	return strings.HasPrefix(base, ".")
}

func readOverride(path string) (*Override, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fs.ErrNotExist
		}
		return nil, fmt.Errorf("read override: %w", err)
	}

	var override Override
	if err := yaml.Unmarshal(data, &override); err != nil {
		return nil, &OverrideError{Path: path, Reason: fmt.Sprintf("parse error: %v", err)}
	}
	if strings.TrimSpace(override.Reason) == "" {
		return nil, &OverrideError{Path: path, Reason: "reason must be non-empty"}
	}
	return &override, nil
}

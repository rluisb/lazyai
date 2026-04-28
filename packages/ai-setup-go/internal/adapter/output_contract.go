// Package adapter — output_contract.go validates the producer/consumer
// chain declared by skill and agent frontmatter (Spec 022 / E2.2).
//
// Each skill's frontmatter declares:
//
//   - output:       canonical artifact path it writes to
//   - consumes:     list of upstream artifacts it reads
//   - produces_for: list of downstream skill / agent names that consume it
//   - mcp_tools:    MCP servers required to run it
//
// At compile time we read every skill / agent file, parse the four fields,
// and check that the declared chain is internally consistent:
//
//   - every name in `produces_for` resolves to a registered skill or agent
//     (otherwise the skill claims a downstream that doesn't exist);
//   - every entry in `consumes` is either a real artifact path under the
//     library or a name produced by *some* upstream skill (otherwise the
//     skill reads from an artifact nobody writes).
//
// The validator emits ContractIssue records — Severity decides whether
// `ai-setup compile` warns or errors.
package adapter

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/frontmatter"
)

// ContractSeverity classifies a contract issue.
type ContractSeverity string

const (
	ContractSeverityWarn  ContractSeverity = "warn"
	ContractSeverityError ContractSeverity = "error"
)

// ContractIssue records a single chain-validation finding.
type ContractIssue struct {
	// Source is the library-relative path of the file the issue belongs to
	// (e.g. "skills/speckit-plan.md"). Empty when the issue isn't tied to a
	// specific file.
	Source string
	// Severity decides whether compile warns or fails.
	Severity ContractSeverity
	// Code is a short stable identifier, e.g. "missing-producer".
	Code string
	// Message is a human-readable description.
	Message string
}

// SkillContract is the parsed subset of skill / agent frontmatter relevant
// to chain validation. All fields are normalized to slices for easier
// comparison; missing fields produce nil slices.
type SkillContract struct {
	Source      string   // library-relative path (e.g. "skills/foo.md")
	Name        string   // frontmatter `name`
	Output      string   // frontmatter `output`
	Consumes    []string // frontmatter `consumes`
	ProducesFor []string // frontmatter `produces_for`
	MCPTools    []string // frontmatter `mcp_tools`
}

// LoadSkillContracts walks libFS reading skills/*.md and agents/*.md,
// returning a SkillContract per file that has frontmatter. Files without
// frontmatter are skipped silently — they aren't part of the contract.
func LoadSkillContracts(libFS fs.FS) ([]SkillContract, error) {
	if libFS == nil {
		return nil, fmt.Errorf("LoadSkillContracts: libFS is nil")
	}

	var contracts []SkillContract
	for _, dir := range []string{"skills", "agents"} {
		entries, err := fs.ReadDir(libFS, dir)
		if err != nil {
			// Missing directory is fine in minimal-library tests.
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".md") {
				continue
			}
			rel := dir + "/" + name
			data, err := fs.ReadFile(libFS, rel)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", rel, err)
			}
			fm, _, err := frontmatter.ExtractFrontmatter(data)
			if err != nil || len(fm) == 0 {
				continue
			}
			c := parseSkillContract(rel, fm)
			contracts = append(contracts, c)
		}
	}
	sort.Slice(contracts, func(i, j int) bool { return contracts[i].Source < contracts[j].Source })
	return contracts, nil
}

// parseSkillContract pulls the relevant fields off a frontmatter map.
func parseSkillContract(source string, fm map[string]any) SkillContract {
	c := SkillContract{Source: source}

	if name, ok := fm["name"].(string); ok {
		c.Name = strings.TrimSpace(name)
	}
	if out, ok := fm["output"].(string); ok {
		c.Output = strings.TrimSpace(out)
	}
	c.Consumes = stringSliceFromFM(fm["consumes"])
	c.ProducesFor = stringSliceFromFM(fm["produces_for"])
	c.MCPTools = stringSliceFromFM(fm["mcp_tools"])
	return c
}

// stringSliceFromFM normalizes a yaml list / scalar / nil into []string.
// Single scalars become a one-element slice; lists are flattened to their
// string elements; everything else returns nil.
func stringSliceFromFM(v any) []string {
	switch t := v.(type) {
	case nil:
		return nil
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil
		}
		return []string{s}
	case []any:
		var out []string
		for _, raw := range t {
			s, ok := raw.(string)
			if !ok {
				continue
			}
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// ValidateContracts cross-checks the loaded contracts and returns any issues
// found. The function is read-only — it never writes to disk.
func ValidateContracts(contracts []SkillContract) []ContractIssue {
	var issues []ContractIssue

	known := map[string]bool{}        // skill / agent name → exists
	produced := map[string][]string{} // produced artifact name → producers
	for _, c := range contracts {
		if c.Name == "" {
			issues = append(issues, ContractIssue{
				Source:   c.Source,
				Severity: ContractSeverityWarn,
				Code:     "missing-name",
				Message:  fmt.Sprintf("%s has no frontmatter `name`; chain edges that target it cannot resolve", c.Source),
			})
			continue
		}
		known[c.Name] = true
		// Every skill produces its own name (downstream skills consume by name).
		produced[c.Name] = append(produced[c.Name], c.Name)
		// Producers also "produce" their declared output path so consumes
		// referencing the literal path resolves cleanly.
		if c.Output != "" {
			produced[c.Output] = append(produced[c.Output], c.Name)
		}
	}

	for _, c := range contracts {
		// produces_for must reference a known skill / agent name.
		for _, downstream := range c.ProducesFor {
			if isPathLikeRef(downstream) {
				continue // path-like targets aren't required to be skill names
			}
			if !known[downstream] {
				issues = append(issues, ContractIssue{
					Source:   c.Source,
					Severity: ContractSeverityError,
					Code:     "missing-downstream",
					Message:  fmt.Sprintf("%s declares produces_for: %q but no such skill or agent exists", c.Name, downstream),
				})
			}
		}
		// consumes must be a known producer or a path-like reference.
		for _, upstream := range c.Consumes {
			if isPathLikeRef(upstream) {
				continue
			}
			if _, ok := produced[upstream]; !ok {
				issues = append(issues, ContractIssue{
					Source:   c.Source,
					Severity: ContractSeverityWarn,
					Code:     "missing-producer",
					Message:  fmt.Sprintf("%s consumes %q but no skill/agent declares output for it", c.Name, upstream),
				})
			}
		}
	}

	return issues
}

// isPathLikeRef reports whether a consumes / produces_for entry looks like
// a literal path (contains "/" or "." or starts with ".specify"). Contract
// validation skips path-like refs because they're real files, not skill
// names — they live on disk and their existence is checked at install time.
func isPathLikeRef(s string) bool {
	if s == "" {
		return false
	}
	if strings.ContainsAny(s, "/.") {
		return true
	}
	return strings.HasPrefix(s, ".")
}

// FormatContractIssues returns a human-friendly multi-line summary of the
// issues, grouped by severity. Returns empty string when issues is empty.
func FormatContractIssues(issues []ContractIssue) string {
	if len(issues) == 0 {
		return ""
	}
	var errors, warns []ContractIssue
	for _, i := range issues {
		if i.Severity == ContractSeverityError {
			errors = append(errors, i)
		} else {
			warns = append(warns, i)
		}
	}
	var b strings.Builder
	if len(errors) > 0 {
		fmt.Fprintf(&b, "contract errors: %d\n", len(errors))
		for _, i := range errors {
			fmt.Fprintf(&b, "  ✗ [%s] %s — %s\n", i.Code, i.Source, i.Message)
		}
	}
	if len(warns) > 0 {
		fmt.Fprintf(&b, "contract warnings: %d\n", len(warns))
		for _, i := range warns {
			fmt.Fprintf(&b, "  ! [%s] %s — %s\n", i.Code, i.Source, i.Message)
		}
	}
	return b.String()
}

// HasContractErrors reports whether the issues list contains at least one
// error-severity entry. Used by `ai-setup compile` to decide whether to
// block.
func HasContractErrors(issues []ContractIssue) bool {
	for _, i := range issues {
		if i.Severity == ContractSeverityError {
			return true
		}
	}
	return false
}

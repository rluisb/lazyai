package compiler

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
)

// ContractSeverity classifies a contract validation issue.
type ContractSeverity string

const (
	ContractSeverityWarn  ContractSeverity = "warn"
	ContractSeverityError ContractSeverity = "error"
)

// ContractIssue records one skill/agent contract validation finding.
type ContractIssue struct {
	Source   string
	Severity ContractSeverity
	Code     string
	Message  string
}

// SkillContract is the normalized subset of skill/agent frontmatter used to
// validate producer/consumer chains.
type SkillContract struct {
	Source      string
	Name        string
	Output      string
	Consumes    []string
	ProducesFor []string
	MCPTools    []string
}

// LoadSkillContracts walks the library skills and agents directories and
// returns contract records for markdown files with a frontmatter name.
func LoadSkillContracts(libFS fs.FS) ([]SkillContract, error) {
	if libFS == nil {
		return nil, fmt.Errorf("LoadSkillContracts: libFS is nil")
	}

	var contracts []SkillContract
	for _, dir := range []string{"skills", "agents"} {
		entries, err := fs.ReadDir(libFS, dir)
		if err != nil {
			// Directory may not exist in minimal libraries.
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			base := entry.Name()
			if base == "AGENTS.md" || base == "AGENT.md" || strings.HasPrefix(base, "_") {
				continue
			}

			source := path.Join(dir, base)
			data, err := fs.ReadFile(libFS, source)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", source, err)
			}

			fm, _, err := frontmatter.ExtractFrontmatter(data)
			if err != nil || len(fm) == 0 {
				fm = extractContractFields(data)
				if len(fm) == 0 {
					continue
				}
			}

			contract := parseSkillContract(source, fm)
			if contract.Name == "" {
				continue
			}
			contracts = append(contracts, contract)
		}
	}

	sort.Slice(contracts, func(i, j int) bool { return contracts[i].Source < contracts[j].Source })
	return contracts, nil
}

func extractContractFields(data []byte) map[string]any {
	// Contract validation only needs a small, flat subset of skill frontmatter.
	// Some library files contain prose-like YAML elsewhere; recover these fields
	// so one malformed metadata section does not make real skill names disappear.
	frontmatterLines, ok := frontmatterLines(data)
	if !ok {
		return nil
	}

	fm := make(map[string]any)
	for i := 0; i < len(frontmatterLines); i++ {
		line := frontmatterLines[i]
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "name", "output":
			if value != "" {
				fm[key] = trimYAMLScalar(value)
			}
		case "consumes", "produces_for", "mcp_tools":
			items := parseContractList(value)
			if value == "" {
				items, i = parseIndentedContractList(frontmatterLines, i)
			}
			if len(items) > 0 {
				fm[key] = items
			}
		}
	}
	return fm
}

func frontmatterLines(data []byte) ([]string, bool) {
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(content, "\n")

	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	if start >= len(lines) || strings.TrimSpace(lines[start]) != "---" {
		return nil, false
	}

	for end := start + 1; end < len(lines); end++ {
		if strings.TrimSpace(lines[end]) == "---" {
			return lines[start+1 : end], true
		}
	}
	return nil, false
}

func parseIndentedContractList(lines []string, keyIndex int) ([]string, int) {
	var items []string
	i := keyIndex + 1
	for ; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			break
		}
		if strings.HasPrefix(trimmed, "-") {
			item := trimYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "-")))
			if item != "" {
				items = append(items, item)
			}
		}
	}
	return items, i - 1
}

func parseContractList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		value = strings.TrimSuffix(strings.TrimPrefix(value, "["), "]")
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := trimYAMLScalar(part)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}

func trimYAMLScalar(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return strings.TrimSpace(value)
}

func parseSkillContract(source string, fm map[string]any) SkillContract {
	contract := SkillContract{Source: source}
	if name, ok := fm["name"].(string); ok {
		contract.Name = strings.TrimSpace(name)
	}
	if output, ok := fm["output"].(string); ok {
		contract.Output = strings.TrimSpace(output)
	}
	contract.Consumes = stringSliceFromFrontmatter(fm["consumes"])
	contract.ProducesFor = stringSliceFromFrontmatter(fm["produces_for"])
	contract.MCPTools = stringSliceFromFrontmatter(fm["mcp_tools"])
	return contract
}

func stringSliceFromFrontmatter(value any) []string {
	switch typed := value.(type) {
	case nil:
		return nil
	case string:
		item := strings.TrimSpace(typed)
		if item == "" {
			return nil
		}
		return []string{item}
	case []string:
		out := make([]string, 0, len(typed))
		for _, raw := range typed {
			if item := strings.TrimSpace(raw); item != "" {
				out = append(out, item)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, raw := range typed {
			item, ok := raw.(string)
			if !ok {
				continue
			}
			if item = strings.TrimSpace(item); item != "" {
				out = append(out, item)
			}
		}
		return out
	default:
		return nil
	}
}

// cleanContractTarget strips parenthetical annotations from produces_for targets.
// Frontmatter entries like "memory-write (if ADR created)" should resolve to
// "memory-write".
func cleanContractTarget(target string) string {
	if idx := strings.Index(target, "("); idx > 0 {
		return strings.TrimSpace(target[:idx])
	}
	return strings.TrimSpace(target)
}

func isContractReference(target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	return !strings.ContainsAny(target, " \t/.")
}

func isExternalContractSink(target string) bool {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "constitution", "memory", "github", "plan-gate":
		return true
	default:
		return false
	}
}

// ValidateChain cross-checks skill/agent contracts using the TypeScript
// compiler contract-validator semantics.
func ValidateChain(contracts []SkillContract) []ContractIssue {
	var issues []ContractIssue
	byName := make(map[string]SkillContract)

	for _, contract := range contracts {
		normalizedName := strings.ToLower(contract.Name)
		if _, exists := byName[normalizedName]; exists {
			issues = append(issues, ContractIssue{
				Source:   contract.Source,
				Severity: ContractSeverityError,
				Code:     "duplicate-name",
				Message:  fmt.Sprintf("skill/agent name %q is declared in multiple files (last: %s)", contract.Name, contract.Source),
			})
		}
		byName[normalizedName] = contract
	}

	producers := make(map[string]string)
	for _, contract := range contracts {
		if contract.Output != "" {
			producers[contract.Output] = contract.Name
		}
	}

	for _, contract := range contracts {
		for _, target := range contract.ProducesFor {
			cleaned := cleanContractTarget(target)
			if !isContractReference(cleaned) || isExternalContractSink(cleaned) {
				continue
			}
			normalizedTarget := strings.ToLower(cleaned)
			if _, exists := byName[normalizedTarget]; !exists {
				issues = append(issues, ContractIssue{
					Source:   contract.Source,
					Severity: ContractSeverityWarn,
					Code:     "missing-downstream",
					Message:  fmt.Sprintf("%q declares produces_for %q, but no skill/agent with that name exists", contract.Name, target),
				})
			}
		}

		for _, artifact := range contract.Consumes {
			if !isContractReference(artifact) {
				continue
			}
			if _, exists := producers[artifact]; exists {
				continue
			}
			issues = append(issues, ContractIssue{
				Source:   contract.Source,
				Severity: ContractSeverityWarn,
				Code:     "missing-producer",
				Message:  fmt.Sprintf("%q declares consumes %q, but no skill produces it (may be a scaffold artifact)", contract.Name, artifact),
			})
		}

		if contract.Output != "" && contract.Output == contract.Name {
			issues = append(issues, ContractIssue{
				Source:   contract.Source,
				Severity: ContractSeverityWarn,
				Code:     "self-output",
				Message:  fmt.Sprintf("%q declares output %q which matches its own name — may indicate a misconfiguration", contract.Name, contract.Output),
			})
		}
	}

	consumedBy := make(map[string]bool)
	for _, contract := range contracts {
		for _, target := range contract.ProducesFor {
			cleaned := cleanContractTarget(target)
			if !isContractReference(cleaned) {
				continue
			}
			consumedBy[strings.ToLower(cleaned)] = true
		}
	}

	rootSkills := map[string]bool{
		"speckit-constitution":          true,
		"rpi":                           true,
		"bugfix":                        true,
		"spike":                         true,
		"proof-of-concept":              true,
		"housekeeping":                  true,
		"implement":                     true,
		"iterate":                       true,
		"orchestrate":                   true,
		"parallel-execution":            true,
		"plan":                          true,
		"research":                      true,
		"self-improve":                  true,
		"memory-write":                  true,
		"tdd-loop":                      true,
		"diagnose":                      true,
		"anti-speculation":              true,
		"chain-verify":                  true,
		"review":                        true,
		"red-team-plan":                 true,
		"github-pr-review":              true,
		"impact-check":                  true,
		"populate":                      true,
		"dynamic-compose":               true,
		"chain-customize":               true,
		"catalog-manage":                true,
		"speckit-implement":             true,
		"improve-codebase-architecture": true,
		"extract-standards":             true,
		"process-audit":                 true,
	}
	for _, contract := range contracts {
		if !strings.HasPrefix(contract.Source, "skills/") {
			continue
		}
		normalizedName := strings.ToLower(contract.Name)
		if consumedBy[normalizedName] || rootSkills[normalizedName] {
			continue
		}
		issues = append(issues, ContractIssue{
			Source:   contract.Source,
			Severity: ContractSeverityWarn,
			Code:     "orphan-skill",
			Message:  fmt.Sprintf("%q is not consumed by any other skill and is not a root skill — may be unreachable", contract.Name),
		})
	}

	return issues
}

// FormatContractIssues returns a human-readable contract validation summary.
func FormatContractIssues(issues []ContractIssue) string {
	if len(issues) == 0 {
		return ""
	}

	var errors, warnings []ContractIssue
	for _, issue := range issues {
		if issue.Severity == ContractSeverityError {
			errors = append(errors, issue)
		} else {
			warnings = append(warnings, issue)
		}
	}

	var b strings.Builder
	if len(errors) > 0 {
		fmt.Fprintf(&b, "contract errors: %d\n", len(errors))
		for _, issue := range errors {
			fmt.Fprintf(&b, "  ✗ [%s] %s — %s\n", issue.Code, issue.Source, issue.Message)
		}
	}
	if len(warnings) > 0 {
		fmt.Fprintf(&b, "contract warnings: %d\n", len(warnings))
		for _, issue := range warnings {
			fmt.Fprintf(&b, "  ! [%s] %s — %s\n", issue.Code, issue.Source, issue.Message)
		}
	}
	return b.String()
}

// HasContractErrors reports whether any issue has error severity.
func HasContractErrors(issues []ContractIssue) bool {
	for _, issue := range issues {
		if issue.Severity == ContractSeverityError {
			return true
		}
	}
	return false
}

// ContractValidationFails applies compile's strictness policy to issues.
func ContractValidationFails(issues []ContractIssue, strict bool) bool {
	return HasContractErrors(issues) || (strict && len(issues) > 0)
}

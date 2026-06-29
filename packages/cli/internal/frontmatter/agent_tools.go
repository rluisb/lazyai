package frontmatter

import (
	"fmt"
	"strings"
)

// AgentToolGrant is a canonical tool-capability token that an agent may be
// granted. The seven tokens below are the only legal values.
type AgentToolGrant string

const (
	AgentToolRead   AgentToolGrant = "read"
	AgentToolEdit   AgentToolGrant = "edit"
	AgentToolShell  AgentToolGrant = "shell"
	AgentToolSearch AgentToolGrant = "search"
	AgentToolWeb    AgentToolGrant = "web"
	AgentToolMCP    AgentToolGrant = "mcp"
	AgentToolSpawn  AgentToolGrant = "spawn"
)

// validToolGrants is the set of canonical tokens, used for O(1) validation.
var validToolGrants = map[AgentToolGrant]struct{}{
	AgentToolRead:   {},
	AgentToolEdit:   {},
	AgentToolShell:  {},
	AgentToolSearch: {},
	AgentToolWeb:    {},
	AgentToolMCP:    {},
	AgentToolSpawn:  {},
}

// ParseAgentToolGrants reads the `tools:` field from agent source frontmatter
// and returns the ordered list of canonical AgentToolGrant values.
//
// It returns (nil, nil) — meaning "unrestricted / backward-compatible" — in
// any of the following cases:
//   - the source has no frontmatter block
//   - the frontmatter has no `tools` key
//   - the `tools` value is present but empty
//
// It returns a non-nil error if any token is not one of the seven canonical
// values (read, edit, shell, search, web, mcp, spawn).
//
// Input order is preserved. Duplicate tokens are passed through without
// deduplication; callers that need a set should deduplicate themselves.
func ParseAgentToolGrants(source []byte) ([]AgentToolGrant, error) {
	if !HasFrontmatter(source) {
		return nil, nil
	}

	fm, _, err := ExtractFrontmatter(source)
	if err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	raw, ok := fm["tools"]
	if !ok || raw == nil {
		return nil, nil
	}

	var tokens []string
	switch v := raw.(type) {
	case []any:
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("invalid tool grant %v: tools entries must be strings", item)
			}
			tokens = append(tokens, s)
		}
	case string:
		tokens = append(tokens, strings.FieldsFunc(v, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
		})...)
	default:
		return nil, fmt.Errorf("invalid tools field: expected string or list of strings")
	}

	if len(tokens) == 0 {
		return nil, nil
	}

	grants := make([]AgentToolGrant, 0, len(tokens))
	for _, tok := range tokens {
		g := AgentToolGrant(tok)
		if _, valid := validToolGrants[g]; !valid {
			return nil, fmt.Errorf("unknown tool grant %q: must be one of read, edit, shell, search, web, mcp, spawn", tok)
		}
		grants = append(grants, g)
	}
	return grants, nil
}

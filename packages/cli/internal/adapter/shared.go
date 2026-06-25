// Package adapter provides shared helper functions used by all tool adapters.
// Ported from the TypeScript shared.ts utilities.
package adapter

import (
	"fmt"
	"path/filepath"
)

// Managed block markers for idempotent root AGENTS.md patching.
const (
	ManagedBlockStartMarker = "<!-- lazyai:managed:start root-agents v1 -->"
	ManagedBlockEndMarker   = "<!-- lazyai:managed:end root-agents v1 -->"
)

// vibe-lab managed-marker contract. Must match bin/inject exactly.
func managedAgentMarker(surface, name string) string {
	return fmt.Sprintf("<!-- vibe-lab:managed kind=agent surface=%s name=%s source=.agents/agents/%s.md -->", surface, name, name)
}

// selectionSet returns a set (map[T]bool) for the given slice, or nil if the
// slice is empty (meaning "install everything").
func selectionSet[T ~string](items []T) map[T]bool {
	if len(items) == 0 {
		return nil
	}
	m := make(map[T]bool, len(items))
	for _, item := range items {
		m[item] = true
	}
	return m
}

const (
	defaultAgentID          = "guide"
	defaultAgentDescription = "Front-door default agent. Answers directly, chats naturally, and only suggests or delegates specialists when that improves the outcome."
)

var canonicalAgentIDs = map[string]struct{}{
	defaultAgentID:      {},
	"implementer":       {},
	"researcher":        {},
	"deployer":          {},
	"responder":         {},
	"planner":           {},
	"reviewer":          {},
	"evidence-verifier": {},
}

func isCanonicalAgentFile(file string) bool {
	_, ok := canonicalAgentIDs[fileID(file)]
	return ok
}

// isPiSystemPromptFile reports whether a library file is one of Pi's
// project system-prompt files: SYSTEM.md (replaces the default prompt) or
// APPEND_SYSTEM.md (appends to the default). Both live at the .pi root and
// are distinct from AGENTS.md / CLAUDE.md context files.
func isPiSystemPromptFile(file string) bool {
	base := filepath.Base(file)
	return base == "SYSTEM.md" || base == "APPEND_SYSTEM.md"
}

func isDefaultAgentFile(file string) bool {
	return fileID(file) == defaultAgentID
}

// Package auth detects which AI-coding-CLI providers the user is currently
// authenticated against. The setup wizard uses these probes to populate the
// OpenCode provider-selection prompt: only providers that pass their probe
// are presented as options, and the user's selection becomes the
// ConfiguredProviders input to models.Resolve.
//
// Probes are deliberately cheap and best-effort: each runs a single
// non-mutating command with a short timeout and treats exit-zero as a green
// light. False negatives (probe fails but provider is actually usable)
// surface as a missing option in the wizard — the user can override
// manually. False positives (probe passes but provider rejects later)
// surface as a runtime error in the AI CLI, not a lazyai compile error.
package auth

import (
	"context"
	"os/exec"
	"sync"
	"time"
)

// ProviderID values match the keys used by models.dev and by
// models.Catalog.DenyProviders / ConfiguredProviders. Adding a new provider
// here means adding a Probe entry and threading the ID through the catalog.
type ProviderID string

const (
	ProviderAnthropic     ProviderID = "anthropic"
	ProviderOpenAI        ProviderID = "openai"
	ProviderGitHubCopilot ProviderID = "github-copilot"
	ProviderOllamaCloud   ProviderID = "ollama-cloud"
	ProviderGoogle        ProviderID = "google"
	ProviderOpenCode      ProviderID = "opencode"
)

// Provider describes one CLI's auth probe. Probe is invoked with a bounded
// context and must return true only when the CLI is authenticated against
// the named provider — being installed is not enough.
type Provider struct {
	ID    ProviderID
	Label string
	Probe func(context.Context) bool
}

// probeTimeout caps each probe individually. 3s is enough for `gh auth
// status` over a slow connection but short enough that DetectAll completes
// well under 30s even when every probe times out.
const probeTimeout = 3 * time.Second

// Probes is the registry consulted by DetectAll. Order is informational —
// callers should not assume it.
var Probes = []Provider{
	{ProviderAnthropic, "Anthropic (Claude Code)", probeClaude},
	{ProviderOpenAI, "OpenAI (Codex CLI)", probeCmd("codex", "auth", "status")},
	{ProviderGitHubCopilot, "GitHub Copilot", probeGitHubCopilot},
	{ProviderOllamaCloud, "Ollama Cloud", probeCmd("ollama", "whoami")},
	{ProviderGoogle, "Google (Gemini CLI)", probeCmd("gemini", "auth", "list")},
	{ProviderOpenCode, "OpenCode bundled", probeCmd("opencode", "auth", "list")},
}

// DetectAll runs every probe in parallel, capped per-probe by probeTimeout.
// Returns the IDs that passed, preserving Probes order in the output.
func DetectAll(parent context.Context) []ProviderID {
	results := make([]bool, len(Probes))
	var wg sync.WaitGroup
	for i, p := range Probes {
		wg.Add(1)
		go func(i int, p Provider) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(parent, probeTimeout)
			defer cancel()
			results[i] = p.Probe(ctx)
		}(i, p)
	}
	wg.Wait()

	detected := make([]ProviderID, 0, len(Probes))
	for i, ok := range results {
		if ok {
			detected = append(detected, Probes[i].ID)
		}
	}
	return detected
}

// probeCmd returns a Probe that runs `name args...` and reports exit-zero.
// If the binary isn't on PATH, the probe returns false without spawning a
// process — exec.Command would still find it on some systems via shell
// builtins which is not what we want.
func probeCmd(name string, args ...string) func(context.Context) bool {
	return func(ctx context.Context) bool {
		if _, err := exec.LookPath(name); err != nil {
			return false
		}
		return exec.CommandContext(ctx, name, args...).Run() == nil
	}
}

// probeClaude is the Anthropic probe. `claude` has no first-class auth-check
// subcommand, so we use --version as a presence check and treat the user as
// authenticated when the binary runs cleanly. The actual auth check happens
// on first request inside Claude Code itself.
func probeClaude(ctx context.Context) bool {
	if _, err := exec.LookPath("claude"); err != nil {
		return false
	}
	return exec.CommandContext(ctx, "claude", "--version").Run() == nil
}

// probeGitHubCopilot requires both `gh auth status` (logged in to GitHub)
// and `gh extension list` containing the copilot extension. Either alone is
// insufficient: a logged-in gh without the extension can't proxy Copilot
// requests, and the extension without auth fails immediately.
func probeGitHubCopilot(ctx context.Context) bool {
	if _, err := exec.LookPath("gh"); err != nil {
		return false
	}
	if exec.CommandContext(ctx, "gh", "auth", "status").Run() != nil {
		return false
	}
	out, err := exec.CommandContext(ctx, "gh", "extension", "list").Output()
	if err != nil {
		return false
	}
	return containsCopilot(out)
}

func containsCopilot(out []byte) bool {
	for i := 0; i+7 <= len(out); i++ {
		if string(out[i:i+7]) == "copilot" {
			return true
		}
	}
	return false
}

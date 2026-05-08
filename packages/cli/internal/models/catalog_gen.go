// Code generated from https://models.dev/api.json. To refresh, run:
//
//	lazyai models sync
//
// The curated tier lists in catalog.go reference IDs from this slice. The
// sync command verifies that every curated ID still exists upstream; if a
// model is removed from models.dev, sync fails loudly so we can update the
// curation rather than ship a stale reference.
//
// Snapshot: 2026-05 (manual seed; replace with sync output when the command
// lands). Only providers relevant to the three target CLIs are included.

package models

type modelEntry struct {
	Provider   string
	ID         string
	Reasoning  bool
	Multimodal bool
	CtxWindow  int
}

var allModels = []modelEntry{
	// anthropic — Claude Code source-of-truth (aliases resolve through here)
	{"anthropic", "claude-opus-4-7", true, true, 1_000_000},
	{"anthropic", "claude-opus-4-6", true, true, 1_000_000},
	{"anthropic", "claude-opus-4-5", true, true, 200_000},
	{"anthropic", "claude-sonnet-4-6", true, true, 1_000_000},
	{"anthropic", "claude-sonnet-4-5", true, true, 200_000},
	{"anthropic", "claude-haiku-4-5", true, false, 200_000},

	// openai — primary OpenCode provider when user has Codex CLI logged in
	{"openai", "gpt-5.5", true, false, 1_050_000},
	{"openai", "gpt-5.5-pro", true, false, 1_050_000},
	{"openai", "gpt-5.4", true, false, 1_050_000},
	{"openai", "gpt-5.4-mini", true, false, 400_000},
	{"openai", "gpt-5.4-nano", true, false, 400_000},
	{"openai", "gpt-5", true, false, 400_000},
	{"openai", "gpt-5-mini", true, false, 400_000},
	{"openai", "gpt-5-nano", true, false, 400_000},
	{"openai", "gpt-5.1-codex", true, false, 400_000},
	{"openai", "gpt-5.1-codex-mini", true, false, 400_000},
	{"openai", "o3", true, false, 200_000},
	{"openai", "o3-mini", true, false, 200_000},

	// github-copilot — Copilot's curated catalog. Includes Claude IDs that
	// are blocked from OpenCode by DenyNamePatterns but allowed on Copilot.
	{"github-copilot", "claude-opus-4.7", true, false, 144_000},
	{"github-copilot", "claude-opus-4.6", true, false, 144_000},
	{"github-copilot", "claude-opus-4.5", true, false, 160_000},
	{"github-copilot", "claude-sonnet-4.6", true, false, 200_000},
	{"github-copilot", "claude-sonnet-4.5", true, false, 144_000},
	{"github-copilot", "claude-haiku-4.5", true, false, 144_000},
	{"github-copilot", "gpt-5.5", true, false, 400_000},
	{"github-copilot", "gpt-5.4", true, false, 400_000},
	{"github-copilot", "gpt-5.4-mini", true, false, 400_000},
	{"github-copilot", "gpt-5", true, false, 128_000},
	{"github-copilot", "gpt-5-mini", true, false, 264_000},
	{"github-copilot", "gemini-3.1-pro-preview", true, true, 128_000},
	{"github-copilot", "gemini-3-pro-preview", true, true, 128_000},
	{"github-copilot", "gemini-3-flash-preview", true, true, 128_000},
	{"github-copilot", "gemini-2.5-pro", false, true, 128_000},

	// google — direct Gemini for OpenCode users with Gemini CLI auth
	{"google", "gemini-3.1-pro-preview", true, true, 1_048_576},
	{"google", "gemini-3-flash-preview", true, true, 1_048_576},

	// ollama-cloud — offline-friendly OpenCode fallback
	{"ollama-cloud", "gpt-oss:120b", true, false, 131_072},
	{"ollama-cloud", "gpt-oss:20b", true, false, 131_072},
	{"ollama-cloud", "minimax-m2.7", true, false, 196_608},
	{"ollama-cloud", "kimi-k2.6:cloud", true, false, 262_144},
	{"ollama-cloud", "qwen3-coder:480b", false, false, 262_144},
	{"ollama-cloud", "deepseek-v4-pro", true, false, 1_048_576},
	{"ollama-cloud", "glm-4.7", true, false, 202_752},

	// opencode — bundled provider; exposes a curated multi-vendor mix
	// (Claude entries here are filtered out by DenyNamePatterns)
	{"opencode", "gpt-5.5", true, false, 1_050_000},
	{"opencode", "gpt-5.4", true, false, 1_050_000},
	{"opencode", "gpt-5.4-mini", true, false, 400_000},
	{"opencode", "glm-4.7", true, false, 204_800},
	{"opencode", "kimi-k2.6", true, false, 262_144},
	{"opencode", "minimax-m2.7", true, false, 204_800},
	{"opencode", "claude-sonnet-4-6", true, true, 1_000_000}, // listed but DenyNamePatterns blocks it
	{"opencode", "claude-opus-4-7", true, true, 1_000_000},   // ditto

	// opencode-go — curated subset, mostly OSS
	{"opencode-go", "glm-5", true, false, 202_752},
	{"opencode-go", "kimi-k2.6", true, false, 262_144},
	{"opencode-go", "deepseek-v4-pro", true, false, 1_000_000},
	{"opencode-go", "minimax-m2.7", true, false, 204_800},
}

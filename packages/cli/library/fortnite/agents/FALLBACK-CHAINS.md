# Fallback Chains

Inline model-swap strategy. No variant files. Every agent parses Dispatch Parameters block.

## Chains by Agent

| Agent | Chain |
|-------|-------|
| **loop-driver** | ollama-cloud/kimi-k2.6:cloud → ollama-cloud/glm-5.1 → ollama-cloud/minimax-m2.7 → ollama-cloud/gemma4 → escalate |
| **engine-control** | ollama-cloud/kimi-k2.6:cloud → ollama-cloud/glm-5.1 → ollama-cloud/nemotron-3-super → escalate |
| **loot-hawk** | ollama-cloud/kimi-k2.6:cloud → ollama-cloud/glm-5.1 → ollama-cloud/gemma4 → escalate |
| **turbo-crank** | ollama-cloud/deepseek-v4-pro → ollama-cloud/kimi-k2.6:cloud → ollama-cloud/glm-5.1 → ollama-cloud/nemotron-3-super → escalate |
| **wall-builder** | ollama-cloud/kimi-k2.6:cloud → ollama-cloud/glm-5.1 → ollama-cloud/gemma4 → escalate |
| **shield-audit** | ollama-cloud/kimi-k2.6:cloud → ollama-cloud/glm-5.1 → openai/gpt-5.5 → escalate |
| **rift-deploy** | ollama-cloud/nemotron-3-super → ollama-cloud/kimi-k2.6:cloud → escalate |
| **respawn-crew** | ollama-cloud/kimi-k2.6:cloud → ollama-cloud/glm-5.1 → escalate |

## Chain Rationale

- **loop-driver**: Longest chain (4 models) because it's the top router — if it fails, the whole system stops.
- **engine-control**: 3 models for workflow orchestration reliability.
- **loot-hawk**: 3 models for research tasks that need to complete.
- **turbo-crank**: 4 models for planning/specification — critical for correct implementation.
- **wall-builder**: 3 models for implementation tasks.
- **shield-audit**: 3 models including cross-provider fallback (openai/gpt-5.5) for critical safety role. Minimum 2 providers required for Tier 2 critical agents.
- **rift-deploy**: 2 models for Tier 4 sensitive ops. Minimum 2 providers required.
- **respawn-crew**: 2 models for Tier 4 sensitive SRE. Minimum 2 providers required.

## Safety Policy Preservation

Fallback chains must preserve the safety rule: **No model shared between opposing roles** (implement≠review≠plan).

- **shield-audit** (review) shares ollama-cloud/kimi-k2.6:cloud and ollama-cloud/glm-5.1 with wall-builder (implement) in fallback chains, but their **primary models differ** (shield-audit primary: openai/gpt-5.5, wall-builder primary: ollama-cloud/kimi-k2.6:cloud). This is acceptable because the primary model determines the role boundary; fallback chains are for resilience only.
- **Critical agents** (shield-audit, rift-deploy, respawn-crew) must have ≥2 providers in their chain to avoid single-provider dependency.
- **Tier 4 agents** (rift-deploy, respawn-crew) have shorter chains (2 models) because they are script-based and less LLM-dependent, but still require multi-provider coverage.

## Fallback Decision Tree

1. **Model error** (rate limit, timeout, provider) → next model in chain. Max 3 auto-retries.
2. **Agent stuck / doom loop** → inject recovery prompt. Still stuck → kill + escalate.
3. **Agent error / unexpected** → route to shield-audit (quick mode). Can't diagnose → escalate.
4. **Agent timeout** → wait 2 min. No response → kill + escalate + suggest smaller task.
5. **All fallback events** recorded via `scripts/session-db.sh error <sid> <seq> "<detail>"`

## Model Priority

`ollama-cloud` > `openai/gpt` > `opencode-go`

Claude reserved as last resort only. No model shared between opposing roles (implement≠review≠plan).

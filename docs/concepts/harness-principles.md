# Harness Principles & Product Boundary

LazyAI and vibe-lab work together to provide a principled, structured workflow for local AI agents. This document is the single source of truth for the architecture, principles, and boundaries of the system.

## The Boundary

The system is strictly divided into two layers:

1. **LazyAI (The Compiler):**
   - Owns the canonical `.ai/` source directory, schemas, manifests, compilation pipeline, and lockfile validation.
   - Responsible for generating tool-native adapter output (`.opencode`, `.claude`, `.github/copilot-instructions.md`, etc.).
   - Remains a **compile-time asset manager**, never a general runtime or orchestrator.
   - Runtime-adjacent features (like the local database for evaluations) are strictly optional modules.

2. **vibe-lab (The Process/Quality Layer):**
   - Owns the embedded assets that define taste, process, and quality.
   - Provides the workflow instructions (skills, hooks, rubrics) that the local host tools execute.
   - Defines the behavioral standard for local agents (the RPI workflow, human gates).

**Host tools** (OpenCode, Claude Code, GitHub Copilot, Pi, Kiro, OMP, Gemini/Antigravity) are the actual execution engines. LazyAI provides the compiled assets; the host tools run them.

---

## LazyAI Principles

- **Canonical source first:** All configurations and agent definitions live in `.ai/`. Hand-edits to generated files are either preserved (if outside managed regions) or overwritten gracefully via the compiler.
- **Tool-native output:** `lazyai compile` produces exact files the target tools expect natively (no generic shim layers).
- **Compile before execute:** The configuration must be compiled to the host's format before the host executes it.
- **Managed regions and lockfile evidence:** Idempotent writes using `.ai/lock.json` and explicit managed-region markers ensure predictable drift resolution.
- **Host tools execute:** The CLI does not execute agent loops.
- **No framework lock-in:** The compiler output doesn't require a runtime dependency to function.
- **Optional runtime-adjacent state:** Any local database or trace state is opt-in; the core is just text compilation.
- **No hidden orchestration:** No background subagents running via LazyAI.
- **No default RAG core:** Retrieval is handled by the host tools or explicitly imported skills, not a hidden LazyAI server.
- **Human-gated quality:** Nothing is fully autonomous without human-gated boundaries.
- **Trace evidence over vibes:** Validation relies on explicit assertions, trace reviews, and golden testing rather than LLM "vibes."
- **Progressive disclosure:** Complex details are revealed to agents only when needed (e.g., via tiered skill exploration).
- **Adapter honesty over fake parity:** Adapters only emit files for capabilities their host tool actually supports.

---

## vibe-lab Principles

- **RPI workflow:** Research first, Plan explicitly, Implement cleanly.
- **Human gates:** Stop and verify before destructive changes or completing major phases.
- **Anti-slop rules:** Strict rules (like `build-fort` and `shield-wall`) to prevent AI-generated boilerplate or lazy architecture.
- **Clean-code-for-agents:** Output must be simple, structured, and parseable for the next agent, not just for humans.
- **Skill/hook/rubric discipline:** Strict schemas and behavioral guidelines for how skills and hooks are authored.
- **Handoff and memory hygiene:** Agents must summarize state and leave clear context before exiting, ensuring the next session starts effectively.
- **Evidence-verifier behavior:** Any validation step must check concrete evidence (tests, command outputs), not just read the diff and say "looks good."
- **Trace/eval improvement thinking:** Failures lead to invariant captures and tests, constantly improving the baseline.

## Related Documents

These documents define the quality standards that the principles above reference:

- [Skill Quality Guidelines](skill-quality.md) — structural and semantic quality requirements for skills, referenced by the "Anti-slop rules" and "Skill/hook/rubric discipline" principles.
- [Agent Contracts](agent-contracts.md) — contract boundaries for agent definitions, referenced by the "Evidence-verifier behavior" and "Human-gated quality" principles.

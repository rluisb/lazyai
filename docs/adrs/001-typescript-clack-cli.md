# ADR-001: TypeScript CLI with `@clack/prompts`

**Date:** 2026-03-30  
**Status:** Accepted  
**Deciders:** ai-setup maintainers

---

## Context

`ai-setup` is a distributed CLI scaffold that needs to be:

- easy to iterate on quickly
- safe and maintainable as command surface grows (`init`, `add`, `update`, `doctor`, `status`)
- usable both interactively and non-interactively (`--no-interactive`)
- friendly in local terminals and CI logs

We needed to pick a language/runtime and a prompt UX library for the first production version.

## Decision

Use:

1. **TypeScript on Node.js** for CLI implementation
2. **`commander`** for command routing and argument parsing
3. **`@clack/prompts`** for interactive terminal prompts and status messaging

## Rationale

### Why TypeScript + Node.js

- Team familiarity and low onboarding friction.
- Strong static typing for config and manifest contracts (`SetupConfig`, `AiSetupConfig`, `FileRecord`).
- Rich npm ecosystem for CLI tooling and packaging.
- Fast build pipeline with `tsup` and straightforward ESM output.

### Why `@clack/prompts`

- Clean, minimal UX for modern terminals.
- Built-in primitives for intro/outro, select/multiselect/text, and spinner/log messages.
- Works well with non-interactive mode patterns where prompts are bypassed.
- Lower complexity than building/maintaining custom prompt formatting.

### Alternatives considered

- **Plain JavaScript:** faster initial setup, but weaker type guarantees as surface area grows.
- **Go/Rust CLI:** stronger single-binary distribution, but higher implementation/iteration cost for this team and project stage.
- **Inquirer/enquirer/custom prompt layer:** viable, but `@clack/prompts` provided simpler UX primitives with less overhead.

## Consequences

**Positive:**

- Faster feature delivery with compile-time safety.
- Better consistency across command input/output flows.
- Easier testability of command and scaffold modules in current toolchain (Vitest + tsup).

**Negative:**

- Requires Node runtime at execution time.
- ANSI-heavy prompt UX depends on terminal capabilities.
- Potential coupling to `@clack/prompts` API choices for future UX changes.

**Neutral:**

- We continue to support both interactive and flag-driven flows.
- Future packaging strategy (global npm vs other distribution) remains open.

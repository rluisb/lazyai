---
name: chain-customize
description: Override step-level agent, domain, and mode at runtime without creating a new catalog entry.
argument-hint: "[chain-name] [overrides]"
trigger: /chain-customize
phase: meta
---

# Chain Customize Skill

Run an existing chain with per-step overrides — swap agents, inject domain knowledge, or change the execution mode — without creating a new catalog entry. Prevents catalog sprawl when you need the same chain with a different posture.

## When to Use

- "Run the feature chain but use `implementor-senior` instead of `builder`"
- "Run the bugfix chain with the `security` domain on the review step"
- "Run the refactor chain in `junior` mode so it asks before acting"
- Any situation where a static chain is *almost* right but needs one or two adjustments

## When NOT to Use

- The chain is perfect as-is → just run it with `orchestrate`
- The changes are structural (adding/removing steps, changing transitions) → use `dynamic-compose` to create a new chain
- You need this customized version permanently → use `catalog-manage` to register it
- You're changing the entire chain's domain → just pass `domainSkill` to `start_chain` directly

## What Can Be Overridden at Runtime

| Override | Scope | How |
|----------|-------|-----|
| **Agent** | Per step | Via `compose_agent` base parameter |
| **Domain skill** | Per step or entire chain | Via `domainSkill` on `start_chain` or per step via `compose_agent` |
| **Mode skill** | Per step or builder steps only | Via `modeSkill` on `start_chain` or per step via `compose_agent` |
| **Skills** | Per step | Cannot override at runtime — requires new chain version |

Note: Step skills, transitions, and gates are structural. If you need to change those, you need a new chain definition, not a runtime override.

## Override Procedure

### Step 1: Identify the Chain and the Overrides

Before starting, be explicit about what you're changing:

```
Chain: feature
Overrides:
  - Step "implement": agent → implementor-senior (instead of builder)
  - Step "review": domain → security (instead of default)
  - Entire chain: mode → senior
```

### Step 2: Start the Chain with Global Overrides

Use `start_chain` for chain-wide domain and mode injection:

```
start_chain({
  chain: "feature",
  task: "Add rate limiting to the payment API",
  domainSkill: "backend",    // applies to all steps
  modeSkill: "senior",       // applies to builder steps
  context: { ... }
})
```

The `domainSkill` and `modeSkill` parameters on `start_chain` are the **global** layer — they apply to all steps (domain) or builder steps only (mode), per the chain's `domain_skill_injection` and `mode_skill_injection` settings.

### Step 3: Override Per-Step Agent or Domain

When a specific step needs a different agent or domain than the chain default:

1. When the chain advances to that step, **use `compose_agent`** instead of dispatching the step's default agent directly:
   ```
   compose_agent({
     base: "implementor-senior",      // override: use senior instead of builder
     domainSkill: "backend",          // keep domain from chain
     modeSkill: "senior"             // keep mode from chain
   })
   ```
   This merges the base agent with domain and mode layers into a single runtime prompt.

2. Dispatch the composed prompt to the subagent for that step.

3. On completion, call `advance_chain` as normal — the override is transparent to the chain state.

### Step 4: When the Override Needs to Vary by Step

For chains where different steps need different domains:

```
Chain: feature (with security-sensitive review)
Step "research":   compose_agent({ base: "scout",    domainSkill: "backend" })
Step "plan":       compose_agent({ base: "planner",  domainSkill: "backend" })
Step "implement":  compose_agent({ base: "implementor-senior", domainSkill: "backend", modeSkill: "senior" })
Step "review":     compose_agent({ base: "reviewer", domainSkill: "security", modeSkill: "senior" })
Step "document":   compose_agent({ base: "documenter" })  // no domain needed
```

This is the most flexible form: every step gets exactly the posture it needs.

## Common Override Patterns

### Higher-Risk Implementation

```
# When changes touch auth, payments, data integrity, or breaking API changes
Step "implement": agent → implementor-senior, mode → senior
```
The senior mode asks for confirmation on non-trivial decisions. The senior implementor has stronger pre-flight checks.

### Lower-Risk Implementation

```
# When changes are docs, minor fixes, or well-understood patterns
Step "implement": agent → implementor-junior, mode → autonomous
```
The junior mode asks more questions but autonomous means fewer confirmation gates on obvious steps.

### Domain-Specific Review

```
# When the feature touches a specific domain concern
Step "review": domain → security   (for auth/data changes)
Step "review": domain → backend     (for API/persistence changes)
Step "review": domain → frontend    (for UI/accessibility changes)
Step "review": domain → devops      (for infra/deployment changes)
```

### Red-Team Adversarial Step

```
# When security is a top concern, add an adversarial review step
Step "review": agent → red-team, domain → security
```
Note: This only works if the chain's `review` step already uses the `reviewer` agent. If the chain doesn't have a review step at all, you need `dynamic-compose` to add one.

## How `compose_agent` Layering Works

The `compose_agent` MCP call merges three layers:

```
1. Base agent (e.g., "builder")
   ↓
2. Domain skill (e.g., "backend") — injected if the agent is in the domain's `applies_to` list
   ↓
3. Mode skill (e.g., "senior") — injected if the step is a builder step and mode_injection allows it
   ↓
4. Runtime prompt — the merged result used to dispatch the agent
```

The domain and mode skills are only injected when they're applicable:
- **Domain**: only injected for agents listed in the domain's `applies_to` field
- **Mode**: only injected for builder steps (implement, fix, iterate) per the chain's `mode_skill_injection`

When you override the base agent via `compose_agent`, the domain and mode layering still follows these rules — the override changes *who* executes, not *how* domain/mode are applied.

## Override vs. New Chain Decision Tree

```
Is the chain structurally the same (same steps, same transitions)?
  YES → Can you express the change as agent/domain/mode overrides?
         YES → Use chain-customize (this skill)
         NO  → Use dynamic-compose to create a new chain
  NO  → Use dynamic-compose to create a new chain

Do you need this customized version permanently?
  YES → After running it successfully, use catalog-manage to register it
  NO  → Run it as an ephemeral override
```

## Hard Rules

1. **Never override step skills or transitions.** Those are structural. If they're wrong, create a new chain.
2. **Always verify the override agent exists.** If you override to "implementor-senior" but that agent isn't available, the step will fail.
3. **Document what you overrode and why.** When using chain-customize, state: "Running feature chain with implementor-senior (risk: auth changes) and security domain on review (impact: payment data)."
4. **Don't override just to override.** If the default chain is fine, run it as-is. Overrides add cognitive overhead for the human reviewing the run.
5. **Override agents, not skills.** The `skills` field on a step is structural. If a step needs different skills, the chain definition needs updating.

## Anti-Patterns

- Creating a `feature-backend-senior` chain just to swap one agent → use chain-customize instead
- Overriding every step's agent → you're effectively running a different chain; use dynamic-compose
- Overriding domain/mode on steps where the domain doesn't apply (e.g., security domain on a documenter) → no-op at best, confusing at worst
- Running `compose_agent` when no override is needed → unnecessary complexity; dispatch the step's default agent directly

## Integration

- **Primary agent:** Orchestrator
- **MCP tools:** `start_chain` (global domain/mode), `compose_agent` (per-step override), `advance_chain`
- **Triggered by:** user request for chain variants, dynamic-compose (when a chain is "almost" right)
- **Output:** same chain execution with different agent composition
- **See also:** `orchestrate` skill, `dynamic-compose` skill, `catalog-manage` skill
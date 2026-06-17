---
name: improve-codebase-architecture
description: Surface architectural friction and propose deepening opportunities — refactors that turn shallow modules into deep ones. Uses Ousterhout deep-module principle. Trigger when user wants to improve architecture, find refactoring opportunities, or make a codebase more testable and AI-navigable.
trigger: /improve-codebase-architecture
phase: architecture-review
techniques: [chain-of-thought]
output_schema:
  sections:
    - Phase 1: Exploration Findings (shallow modules, tight seams, untested boundaries)
    - Phase 2: Candidate List (numbered, problem/solution/benefits)
    - Phase 3: Grilling Loop Outcome (accepted candidates, rejected with reasons, ADRs created)
consumes:
  - codebase area to analyze
  - existing CONTEXT.md and ADRs (if present)
produces_for:
  - memory-write (if ADR created)
  - planner (grilling outcome)
mcp_tools: [filesystem, ripgrep]
harness:
  feed_forward: [codebase area or module]
  contract: [speckit-review]
  anti_slope: [no-interface-design-yet, no-forced-candidates]
workspace:
  scope: [project]
  reads: [relevant modules, CONTEXT.md, docs/adr/]
  writes: [CONTEXT.md updates, ADR files if created]
---

# Improve Codebase Architecture Skill

## Phase 1 — Explore

Read the project's domain glossary (CONTEXT.md) and any ADRs in the area you're touching first.

Explore the codebase organically. Do NOT follow rigid heuristics. Note where you experience friction:
- Where does understanding one concept require bouncing between many small modules?
- Where are modules shallow — interface nearly as complex as the implementation?
- Where have pure functions been extracted just for testability, but the real bugs hide in how they're called (no locality)?
- Where do tightly-coupled modules leak across their seams?
- Which parts are untested or hard to test through their current interface?

Apply the deletion test to anything you suspect is shallow: would deleting it concentrate complexity, or just move it? A "yes, concentrates" is the signal you want.

## Phase 2 — Present Candidates

Present a numbered list of deepening opportunities. For each candidate include:
- **Files**: which files/modules are involved
- **Problem**: why the current architecture is causing friction
- **Solution**: plain English description of what would change
- **Benefits**: explained in terms of locality and leverage, and also in how tests would improve

Use CONTEXT.md vocabulary for the domain, and the Architecture Vocabulary (below) for the architecture.

**ADR conflicts**: if a candidate contradicts an existing ADR, surface it clearly marked ("contradicts ADR-NNN — but worth reopening because…"). Do NOT silently ignore conflicts.

**Do NOT propose interfaces yet.** Ask the user: "Which of these would you like to explore?"

## Phase 3 — Grilling Loop

Once the user picks a candidate, walk the design tree with them — constraints, dependencies, the shape of the deepened module, what sits behind the seam, what tests survive.

Side effects happen inline as decisions crystallize:
- **Naming a term not in CONTEXT.md?** Add it to CONTEXT.md right there.
- **Sharpening a fuzzy term?** Update CONTEXT.md inline.
- **User rejects a candidate with a load-bearing reason?** Offer an ADR: "Want me to record this as an ADR so future architecture reviews don't re-suggest it?"
- **Want to explore alternative interfaces?** Spawn parallel agents to produce radically different interface designs, then compare.

When done, summarize: which candidates were accepted, which were rejected and why, and any ADRs created.

## Anti-Speculation Rules
- Do NOT propose interfaces in Phase 2 — only surface candidates. Interface design happens in Phase 3 after the user picks a candidate.
- Do NOT force candidates — if no shallow modules found, report that finding and exit cleanly.
- Only offer an ADR when all three are true: (1) hard to reverse, (2) surprising without context, (3) result of a real trade-off.

## Architecture Vocabulary

Glossary of terms used by this skill. Use these terms exactly — consistent language is the point.

### Module
Anything with an interface and an implementation: function, class, package, slice.

### Interface
Everything a caller must know to use the module: types, invariants, error modes, ordering, config. Not just the type signature.

### Implementation
The code inside a module.

### Depth
Leverage at the interface: a lot of behaviour behind a small interface. Deep = high leverage. Shallow = interface nearly as complex as the implementation.

### Seam
Where an interface lives — a place behaviour can be altered without editing in place. Use this, not "boundary."

### Adapter
A concrete thing satisfying an interface at a seam.

### Leverage
What callers get from depth — the benefit of a deep module.

### Locality
What maintainers get from depth: change, bugs, knowledge concentrated in one place.

### Deletion Test
Imagine deleting the module. If complexity vanishes, it was a pass-through (shallow). If complexity reappears across N callers, it was earning its keep (deep). Ask: "does deleting this concentrate or distribute complexity?"

### Interface Is the Test Surface
The interface is what you test — not the implementation details.

### One Adapter = Hypothetical Seam, Two Adapters = Real Seam
One adapter = you could swap the implementation. Two adapters = there is actually a seam you can exploit.

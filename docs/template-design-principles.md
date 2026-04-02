# Template Design Principles

## Why Templates Are Structured This Way

The templates in `library/` are designed to **constrain LLM behavior** for better outcomes. Each structural choice has a specific purpose.

### 1. Prevent Premature Implementation
Templates force agents to complete research and planning phases before writing code. The workflow rule (`rules/workflow.md`) enforces this sequence: Research → Plan → Implement.

### 2. Force Uncertainty Acknowledgment
The uncertainty markers template (`constitution/uncertainty.md`) requires agents to declare confidence levels. This prevents over-confident responses to ambiguous requirements.

### 3. Structured Thinking via Checklists
Quality gates (`constitution/quality-gates.md`) provide concrete verification steps. Agents must check off each gate rather than self-assessing "it looks right."

### 4. Constitutional Compliance
The constitution template establishes project-level principles that override individual agent preferences. This creates consistent behavior across different models and sessions.

### 5. Test-First Ordering
The TDD loop skill (`skills/tdd-loop.md`) structures implementation to write tests before production code. This constrains the implementation to match specifications.

### 6. Scope Containment
The anti-speculation skill (`skills/anti-speculation.md`) and constraint techniques prevent agents from implementing features not explicitly requested.

### 7. Context Management
Token discipline sections and compaction protocols prevent context window exhaustion during long sessions.

## Customization Guidelines

When modifying templates:
- **Preserve the ordering** — phases exist to prevent premature action
- **Keep checklists** — they are verification points, not bureaucracy
- **Maintain uncertainty markers** — removing them degrades output quality
- **Don't remove negative examples** — they prevent common failure modes
- **Keep the "When to Invoke" sections** — they ensure correct agent selection

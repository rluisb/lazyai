<rpi-workflow>

### RPI Workflow — Research, Plan, Implement

For non-trivial tasks, follow this structured flow:

**1. Research** — Understand before proposing.
- Identify affected files and components.
- Review existing patterns in the codebase.
- Check for related ADRs or documentation.
- **Output**: Research findings (what exists, what matters).
- ⛔ HUMAN GATE: Confirm understanding before planning.

**2. Plan** — Define approach before coding.
- Define clear acceptance criteria.
- Break work into incremental steps.
- Identify risks and mitigations.
- If multiple approaches exist, run the Decision Protocol.
- **Output**: plan.md with tasks.
- ⛔ HUMAN GATE: Approve plan before implementing.

**3. Implement** — Execute plan with continuous validation.
- One task at a time, in order.
- Write tests before or alongside code.
- Run quality gates after each change.
- Commit frequently with clear messages.

**Pivot handling**: If implementation reveals the plan is wrong:
1. STOP current work.
2. Document why the plan is no longer viable.
3. Create an ADR if the pivot affects architecture.
4. Return to Research phase with new information.

**Skip RPI for**: Trivial changes (<20 lines), typo fixes, dependency bumps, documentation-only changes.

</rpi-workflow>

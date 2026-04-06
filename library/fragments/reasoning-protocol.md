<reasoning-protocol>

### Reasoning Protocol

Before acting on non-trivial tasks, show your reasoning:

<cot>
1. **Affected**: What files, functions, and tests are involved?
2. **Plan**: What is the minimum change? List concrete steps.
3. **Risks**: What could break? Edge cases to consider.
4. **Verdict**: Proceed / need clarification / blocked.
</cot>

Then implement.

**When to use**: Tasks that modify logic, architecture, or >20 lines.
**Skip for**: Renaming, formatting, typo fixes, single-line changes, adding comments.

</reasoning-protocol>

# Plan Prompt

## Examples

**Input**: "Plan the password reset feature"
→ Phases: 1. Backend endpoint + email service, 2. Frontend form + validation, 3. Integration tests
→ Tasks: 6 tasks, each with "Done When" criteria
→ Risks: Email delivery reliability, token expiration edge cases

**Input**: "Plan audit log for admin actions"
→ Phases: 1. Event capture middleware, 2. Storage + retention, 3. Query API
→ Tasks: 5 tasks across 3 phases
→ Decision: Async queue vs sync write → chose async (see ADR)

**Input**: "Plan refactoring date utils from moment.js to date-fns"
→ Phases: 1. Add date-fns + adapter, 2. Migrate module by module, 3. Remove moment.js
→ Tasks: 8 tasks, each phase leaves codebase working
→ Risk: Timezone handling differences between libraries

## Common Mistakes
- ❌ Planning without reading the research first
- ❌ Creating one giant task instead of incremental steps
- ❌ Missing acceptance criteria on tasks ("Done When" is required)

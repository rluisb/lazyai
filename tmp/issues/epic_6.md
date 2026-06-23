**Epic — RPI Cycle 6: Context/handoff, multi-agent, init & server clarity**

Goal: Complete guidance/clarity around context compaction, handoff, multi-agent usage, init headless populate, and server capability levels — as guidance/templates/honest CLI messaging, never runtime.

### Why
These topics easily overgrow into orchestration. `library/fragments/context-discipline.md` exists; `cmd/init.go`, `cmd/server.go`, `cmd/doctor.go`, `cmd/status.go` exist. Missing: handoff/compaction templates, multi-agent boundary docs, and clarified init/server messaging.

### Tasks (sub-issues)
- Context compaction + handoff templates + concept doc
- Multi-agent boundary docs/templates/fragments
- Init headless populate clarity (help + docs)
- Server L1/L3 capability clarity (cmd + docs + doctor/status)
- Tests for changed help text / docs references

### Boundary
No orchestration command, no runtime scheduler, no fake support claims. Guidance compiles as assets/templates where appropriate.

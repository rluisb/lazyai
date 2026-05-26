---
name: shield-wall
description: Backend anti-slop guardrails. Prevents AI from generating low-quality backend code. 14 anti-patterns across Data Layer, Resilience, Security, and Architecture. Tunable dials for resilience, performance, security, and observability. 6 backend archetypes with preset configurations.
trigger: /shield-wall
triggers:
  - "backend quality check"
  - "review backend code"
  - "backend anti-patterns"
  - "server-side quality"
  - "backend code review"
skill_path: skills/shield-wall
---

## Quick Reference

| | |
|---|---|
| **Use when** | Backend code review, anti-pattern detection, quality gates |
| **Do not use when** | Frontend code (use build-fort), research |
| **Primary agent** | wall-builder |
| **Runtime risk** | Low — advisory guardrails |
| **Outputs** | Anti-pattern reports, severity ratings, checklist |
| **Validation** | Pattern coverage, archetype compliance |
| **Deep mode trigger** | `/shield-wall` or backend quality review |

# Shield Wall — Backend Anti-Slop Guardrails

> *"The storm doesn't care about your deadlines. Your backend shouldn't either."*

## Purpose

Shield Wall is the **fortified perimeter** for all backend code generation in the Fortnite squad. It exists to prevent the most common and costly anti-patterns that turn a solid drop into a lootless elimination.

**What it prevents:**
- Database performance disasters (N+1, missing indexes, unbounded queries)
- Cascading failures (no retries, no circuit breakers, no timeouts)
- Security breaches (hardcoded secrets, unvalidated input, no rate limiting)
- Architectural rot (god objects, tight coupling, blind services)

**When to raise the shield:**
- Before writing any database query or migration
- Before making any external HTTP/gRPC call
- Before exposing any new API endpoint
- Before deploying to staging or production

---

## Anti-Pattern Catalog

### 1. Data Layer — The Foundation

| # | Anti-Pattern | Severity | Why It Kills You | Detection |
|---|-------------|----------|------------------|-----------|
| 1.1 | **N+1 Query** | CRITICAL | O(n) database round-trips in loops. Kills latency at scale. | Look for queries inside loops without eager loading or batching |
| 1.2 | **Missing Index** | HIGH | Full table scans on hot paths. Database CPU spikes. | Check WHERE, JOIN, ORDER BY columns for index coverage |
| 1.3 | **No Pagination** | HIGH | Unbounded result sets. Memory exhaustion. OOM kills. | Verify LIMIT/OFFSET or cursor pagination on list endpoints |
| 1.4 | **Raw SQL Injection** | CRITICAL | Unparameterized queries. Direct string concatenation. | Flag any SQL built with string interpolation |

### 2. Resilience — The Storm Shield

| # | Anti-Pattern | Severity | Why It Kills You | Detection |
|---|-------------|----------|------------------|-----------|
| 2.1 | **No Retry Logic** | HIGH | Transient failures become permanent. False negatives. | Check external calls for retry with exponential backoff |
| 2.2 | **No Circuit Breaker** | MEDIUM | Cascading failures across services. Death spirals. | Verify circuit breaker on all external service calls |
| 2.3 | **No Timeout** | CRITICAL | Hanging requests consume threads. Thread pool exhaustion. | Every external call must have explicit timeout |
| 2.4 | **No Idempotency** | HIGH | Duplicate processing. Double charges. Data corruption. | Check mutation endpoints for idempotency keys |

### 3. Security — The Perimeter

| # | Anti-Pattern | Severity | Why It Kills You | Detection |
|---|-------------|----------|------------------|-----------|
| 3.1 | **Hardcoded Secrets** | CRITICAL | Credentials in repos. Instant breach. Compliance failure. | Scan for API keys, passwords, tokens in source code |
| 3.2 | **No Input Validation** | CRITICAL | Injection attacks. Data corruption. Unexpected behavior. | Verify validation on all user inputs and API parameters |
| 3.3 | **No Rate Limiting** | HIGH | DDoS vulnerability. Resource exhaustion. Abuse. | Check public endpoints for rate limiting middleware |

### 4. Architecture — The Blueprint

| # | Anti-Pattern | Severity | Why It Kills You | Detection |
|---|-------------|----------|------------------|-----------|
| 4.1 | **God Object** | HIGH | Single class knows too much. Untestable. Unmaintainable. | Flag classes >500 lines or >10 dependencies |
| 4.2 | **Tight Coupling** | MEDIUM | Changes ripple. Can't test in isolation. Deployment risk. | Check for direct instantiation of external dependencies |
| 4.3 | **Missing Observability** | MEDIUM | Flying blind. Can't debug. Can't optimize. | Verify metrics, logs, and traces on all services |

---

## Tunable Dials

Configure the shield intensity per project or per service.

| Dial | Values | Default | Description |
|------|--------|---------|-------------|
| **RESILIENCE_LEVEL** | `minimal` / `standard` / `fortress` | `standard` | How aggressively to enforce retries, circuit breakers, timeouts, and idempotency |
| **PERFORMANCE_TARGET** | `lenient` / `balanced` / `strict` | `balanced` | Query complexity limits, pagination requirements, index enforcement |
| **SECURITY_POSTURE** | `internal` / `standard` / `paranoid` | `standard` | Input validation depth, secret scanning, rate limiting thresholds |
| **OBSERVABILITY_DEPTH** | `basic` / `standard` / `deep` | `standard` | Required telemetry: logs, metrics, traces, distributed tracing |

### Dial Presets by Context

| Context | RESILIENCE | PERFORMANCE | SECURITY | OBSERVABILITY |
|---------|-----------|-------------|----------|---------------|
| Internal tool | `minimal` | `lenient` | `internal` | `basic` |
| Standard API | `standard` | `balanced` | `standard` | `standard` |
| Payment service | `fortress` | `strict` | `paranoid` | `deep` |
| Public-facing API | `fortress` | `strict` | `paranoid` | `deep` |

---

## Backend Archetypes

Each archetype comes with **preset dial configurations** and **key concerns**.

### 1. CRUD API
**Purpose:** Standard REST/GraphQL resource API
**Key Concerns:** N+1 queries, pagination, input validation, rate limiting
**Dial Preset:**
```yaml
RESILIENCE_LEVEL: standard
PERFORMANCE_TARGET: balanced
SECURITY_POSTURE: standard
OBSERVABILITY_DEPTH: standard
```

### 2. Event Processor
**Purpose:** Async message/queue consumer
**Key Concerns:** Idempotency, retry logic, dead letter handling, observability
**Dial Preset:**
```yaml
RESILIENCE_LEVEL: fortress
PERFORMANCE_TARGET: balanced
SECURITY_POSTURE: standard
OBSERVABILITY_DEPTH: deep
```

### 3. Auth Service
**Purpose:** Authentication and authorization
**Key Concerns:** Hardcoded secrets, input validation, rate limiting, tight coupling
**Dial Preset:**
```yaml
RESILIENCE_LEVEL: fortress
PERFORMANCE_TARGET: strict
SECURITY_POSTURE: paranoid
OBSERVABILITY_DEPTH: deep
```

### 4. Data Pipeline
**Purpose:** ETL, batch processing, data movement
**Key Concerns:** Missing indexes, no pagination, no observability, god objects
**Dial Preset:**
```yaml
RESILIENCE_LEVEL: standard
PERFORMANCE_TARGET: strict
SECURITY_POSTURE: standard
OBSERVABILITY_DEPTH: deep
```

### 5. Real-time Service
**Purpose:** WebSockets, SSE, real-time notifications
**Key Concerns:** No timeouts, no circuit breakers, missing observability
**Dial Preset:**
```yaml
RESILIENCE_LEVEL: fortress
PERFORMANCE_TARGET: strict
SECURITY_POSTURE: standard
OBSERVABILITY_DEPTH: deep
```

### 6. Batch Job
**Purpose:** Scheduled or triggered background jobs
**Key Concerns:** No idempotency, no retry logic, god objects, tight coupling
**Dial Preset:**
```yaml
RESILIENCE_LEVEL: standard
PERFORMANCE_TARGET: balanced
SECURITY_POSTURE: standard
OBSERVABILITY_DEPTH: standard
```

---

## Gate Functions

Raise the shield at these **three critical checkpoints**:

### Gate 1: Before Writing a Database Query

```
□ Is this query inside a loop? → Use eager loading or batching
□ Are WHERE/JOIN columns indexed? → Add index if missing
□ Is there a LIMIT clause? → Add pagination
□ Is user input parameterized? → Never use string interpolation
□ Is the query explainable? → Run EXPLAIN on hot paths
```

### Gate 2: Before Making an External Call

```
□ Is there a timeout? → Set explicit timeout (default: 5s)
□ Is there retry logic? → Add exponential backoff (max 3 retries)
□ Is there a circuit breaker? → Break after 5 failures in 60s
□ Is the call idempotent? → Add idempotency key for mutations
□ Is the response validated? → Schema validation before processing
```

### Gate 3: Before Exposing an Endpoint

```
□ Is input validated? → Schema validation on all inputs
□ Is there rate limiting? → Add per-user and per-IP limits
□ Are secrets externalized? → No hardcoded keys, use vault/ENV
□ Is authentication required? → Verify auth middleware
□ Is there observability? → Add metrics, logs, and traces
```

---

## Integration

Shield Wall integrates with the full Fortnite squad:

| Skill | Integration Point | How They Work Together |
|-------|-------------------|------------------------|
| **build-mode** | Pre-write gate | Shield Wall runs before build-mode writes backend code. If a gate fails, build-mode must fix before proceeding |
| **zero-point** | Post-write verification | After build-mode completes, zero-point runs Shield Wall checks as part of its quality gate |
| **drift-scope** | Spec alignment | drift-scope verifies that backend implementations still pass Shield Wall gates over time |
| **truth-chain** | Audit trail | Every Shield Wall gate pass/fail is logged to the immutable ledger for compliance and debugging |

### Integration Flow

```
build-mode (write) → Shield Wall (gate check) → zero-point (verify) → truth-chain (log)
                                      ↓
                              drift-scope (periodic re-check)
```

---

## Rules

### Mandatory Enforcement

1. **CRITICAL severity anti-patterns are blockers.** Code with CRITICAL findings cannot be committed or deployed.
2. **HIGH severity anti-patterns require justification.** Document why the risk is acceptable, or fix before merge.
3. **MEDIUM severity anti-patterns are warnings.** Must be tracked; should be fixed in the next sprint.
4. **All database queries must pass Gate 1.** No exceptions.
5. **All external calls must pass Gate 2.** No exceptions.
6. **All new endpoints must pass Gate 3.** No exceptions.
7. **Dial presets are defaults, not suggestions.** Override only with explicit approval.
8. **Archetype selection is mandatory.** Every backend service must declare its archetype.

### Override Protocol

```
1. Document the override reason in the PR description
2. Get approval from a senior engineer
3. Log the override to truth-chain
4. Set a remediation date (max 30 days)
```

---

## Usage

### Trigger the Shield

```
/shield-wall
```

### Or use natural language

- "Run backend quality check on this service"
- "Review backend code for anti-patterns"
- "Check server-side quality before merge"
- "Backend code review for the auth module"

### In Build Mode

When build-mode is generating backend code, it automatically invokes Shield Wall at each gate. If a gate fails, build-mode receives:

```
[SHIELD WALL BLOCKED] Gate 1: N+1 Query detected in UserController.list()
[REQUIRED] Use eager loading or batch query
```

---

## Storm Rating

| Metric | Value |
|--------|-------|
| Anti-patterns covered | 14 |
| Severity levels | CRITICAL / HIGH / MEDIUM |
| Tunable dials | 4 |
| Backend archetypes | 6 |
| Gate checkpoints | 3 |
| Integration skills | 4 |

**Shield Wall Status:** 🛡️ **RAISED** — No slop gets through.

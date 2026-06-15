---
name: qmd
description: Search markdown knowledge bases, notes, and documentation using QMD. Use when users ask to search notes, find documents, or look up information. Supports hybrid BM25 + vector search.
trigger: /qmd
license: MIT
compatibility: Requires qmd CLI (npm, NOT bun). Install via `npm install -g @tobilu/qmd`. Vector search requires npm version (bun sqlite lacks vec0 support).
metadata:
  author: tobi
  version: "2.1.0"
  skill_path: skills/qmd/
  scripts: []
  database_schema: "qmd index at ~/.cache/qmd/index.sqlite — 2382 docs, 5558+ vectors, second-brain collection"
allowed-tools: Bash(qmd:*), mcp__qmd__*
---

# QMD - Quick Markdown Search

Local search engine for markdown content.

## Status

!`qmd status 2>/dev/null || echo "Not installed: npm install -g @tobilu/qmd"`

## MCP: `query`

```json
{
  "searches": [
    { "type": "lex", "query": "CAP theorem consistency" },
    { "type": "vec", "query": "tradeoff between consistency and availability" }
  ],
  "collections": ["docs"],
  "limit": 10
}
```

### Query Types

| Type | Method | Input |
|------|--------|-------|
| `lex` | BM25 | Keywords — exact terms, names, code |
| `vec` | Vector | Question — natural language |
| `hyde` | Vector | Answer — hypothetical result (50-100 words) |

### Writing Good Queries

**lex (keyword)**
- 2-5 terms, no filler words
- Exact phrase: `"connection pool"` (quoted)
- Exclude terms: `performance -sports` (minus prefix)
- Code identifiers work: `handleError async`

**vec (semantic)**
- Full natural language question
- Be specific: `"how does the rate limiter handle burst traffic"`
- Include context: `"in the payment service, how are refunds processed"`

**hyde (hypothetical document)**
- Write 50-100 words of what the *answer* looks like
- Use the vocabulary you expect in the result

**expand (auto-expand)**
- Use a single-line query (implicit) or `expand: question` on its own line
- Lets the local LLM generate lex/vec/hyde variations
- Do not mix `expand:` with other typed lines — it's either a standalone expand query or a full query document

### Intent (Disambiguation)

When a query term is ambiguous, add `intent` to steer results:

```json
{
  "searches": [
    { "type": "lex", "query": "performance" }
  ],
  "intent": "web page load times and Core Web Vitals"
}
```

Intent affects expansion, reranking, chunk selection, and snippet extraction. It does not search on its own — it's a steering signal that disambiguates queries like "performance" (web-perf vs team health vs fitness).

### Combining Types

| Goal | Approach |
|------|----------|
| Know exact terms | `lex` only |
| Don't know vocabulary | Use a single-line query (implicit `expand:`) or `vec` |
| Best recall | `lex` + `vec` |
| Complex topic | `lex` + `vec` + `hyde` |
| Ambiguous query | Add `intent` to any combination above |

First query gets 2x weight in fusion — put your best guess first.

### Lex Query Syntax

| Syntax | Meaning | Example |
|--------|---------|---------|
| `term` | Prefix match | `perf` matches "performance" |
| `"phrase"` | Exact phrase | `"rate limiter"` |
| `-term` | Exclude | `performance -sports` |

Note: `-term` only works in lex queries, not vec/hyde.

### Collection Filtering

```json
{ "collections": ["docs"] }              // Single
{ "collections": ["docs", "notes"] }     // Multiple (OR)
```

Omit to search all collections.

## Other MCP Tools

| Tool | Use |
|------|-----|
| `get` | Retrieve doc by path or `#docid` |
| `multi_get` | Retrieve multiple by glob/list |
| `status` | Collections and health |

## CLI

```bash
qmd query "question"              # Auto-expand + rerank
qmd query $'lex: X\nvec: Y'       # Structured
qmd query $'expand: question'     # Explicit expand
qmd query --json --explain "q"    # Show score traces (RRF + rerank blend)
qmd search "keywords"             # BM25 only (no LLM)
qmd get "#abc123"                 # By docid
qmd multi-get "journals/2026-*.md" -l 40  # Batch pull snippets by glob
qmd multi-get notes/foo.md,notes/bar.md   # Comma-separated list, preserves order
```

## HTTP API

```bash
curl -X POST http://localhost:8181/query \
  -H "Content-Type: application/json" \
  -d '{"searches": [{"type": "lex", "query": "test"}]}'
```

## Setup

```bash
npm install -g @tobilu/qmd
qmd collection add ~/notes --name notes
qmd embed
```

## Agent Wiring

qmd is wired to these agents for vault research:

| Agent | When to use qmd |
|-------|----------------|
| **loot-hawk** | Primary vault research tool — use `qmd query` with hybrid lex+vec for deep exploration |
| **turbo-crank** | Phase 0 (Grill Me With Docs) — search vault for existing knowledge before asking questions |
| **shield-audit** | Security/adversarial mode — search vault for known vulnerability patterns, threat models |
| **loop-driver** | Routing decisions — search vault for architectural patterns, past decisions |

### Recommended Query Patterns per Agent

**loot-hawk (deep research):**
```bash
qmd query
lex: exact-terms\nvec: natural language question about the topic' -c second-brain -l 10
```

**turbo-crank (Grill Me With Docs):**
```bash
qmd query
lex: domain-keywords\nvec: what are the key decisions and tradeoffs for {topic}' -c second-brain -l 5
```

**shield-audit (security review):**
```bash
qmd query
lex: vulnerability-pattern\nvec: how could this be exploited or misused' -c second-brain -l 5
```

**loop-driver (routing context):**
```bash
qmd search "keyword" -c second-brain -l 3
```

## Important Notes

- **MUST use npm version** — bun's sqlite does not support vec0 module for vector search
- **842 docs pending embeddings** — run `qmd embed` to complete vector index for full recall
- **Hybrid search (lex+vec)** gives best results — lex for exact terms, vec for semantic meaning
- **Reranking** is automatic and uses local Qwen3-Reranker-0.6B model (runs on Apple M3 Metal GPU)

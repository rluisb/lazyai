# PRD: AI Setup CLI

**Feature:** 001-cli-scaffold
**Date:** 2026-03-28
**Status:** Draft
**Research:** [research.md](./research.md)

---

## Problem Statement

Teams adopting our AI-assisted development setup must manually copy 29+ files, create 15+ directories, adapt content per AI tool, and fill dozens of placeholders. This takes hours, is error-prone, and leads to inconsistent setups across repos. New projects and new team members face a high barrier to entry.

## Goals

- G1: One command sets up the complete AI development structure in any repo
- G2: Support multiple AI tools (starting with Pi and OpenCode) from the same source
- G3: Workspace mode for orgs managing multiple repos with shared conventions
- G4: Update command that refreshes library files without overwriting customizations
- G5: Doctor command that verifies setup integrity

## Non-Goals

| What | Why |
|------|-----|
| Runtime AI agent orchestration | This is a scaffold tool, not an agent framework |
| MCP server management | Out of scope — tools handle their own MCP setup |
| Cloud/SaaS features | Local-only, repo-contained. Zero external dependencies at runtime. |
| GUI or web interface | CLI only. Terminal-native. |
| Supporting all tools at once | Start with Pi + OpenCode. Others added via adapter pattern later. |

## Existing Landscape

- `npx skills add` — installs skills only, not full project structure
- `specify init` — spec-kit specific, tied to their ecosystem
- Our `AI-Agentic-Setup-Templates/` — the 29 files, but no installer

## User Stories

### US-1: Init a project (P1) ⭐ MVP

**As a** developer, **I want** to run one command to scaffold the AI setup in my repo **so that** I start with the full structure, templates, agents, and rules without manual copying.

**Acceptance Criteria:**
1. GIVEN I'm in a git repo WHEN I run `npx @org/ai-setup init` THEN I see an interactive prompt asking setup type and tools
2. GIVEN I select "Project" and "Pi + OpenCode" WHEN setup completes THEN all docs/ structure exists with AGENTS.md context files, templates, rules
3. GIVEN I selected Pi WHEN setup completes THEN `.pi/agents/`, `.pi/skills/`, `.pi/templates/` exist with correct content
4. GIVEN I selected OpenCode WHEN setup completes THEN `.opencode/agents/`, `.opencode/commands/` exist with correct content
5. GIVEN setup completes WHEN I open AGENTS.md THEN it contains `[YOUR_*]` placeholders and the decision tree

### US-2: Init a workspace (P2)

**As a** tech lead, **I want** to set up a workspace-level config for multiple repos **so that** all projects inherit shared conventions.

**Acceptance Criteria:**
1. GIVEN I'm in a directory with repos WHEN I run init and select "Workspace" THEN a `.ai-workspace/` directory is created with shared rules and standards
2. GIVEN a workspace exists WHEN I init a project inside it THEN the project inherits workspace conventions

### US-3: Add a tool (P2)

**As a** developer, **I want** to add support for a new AI tool to an existing setup **so that** I don't have to re-init the whole project.

**Acceptance Criteria:**
1. GIVEN a project with Pi setup WHEN I run `npx @org/ai-setup add opencode` THEN OpenCode-specific files are added without touching existing files

### US-4: Update library files (P2)

**As a** developer, **I want** to update my setup when the library improves **so that** I get new templates and improvements without losing my customizations.

**Acceptance Criteria:**
1. GIVEN files I haven't customized WHEN I run update THEN they are overwritten with latest
2. GIVEN files I have customized WHEN I run update THEN they are skipped with a warning
3. GIVEN new library files WHEN I run update THEN they are added

### US-5: Check setup health (P3)

**As a** developer, **I want** to verify my setup is complete and consistent **so that** I catch problems before they affect my work.

**Acceptance Criteria:**
1. GIVEN a setup WHEN I run `npx @org/ai-setup doctor` THEN it reports: missing files, AGENTS.md/CLAUDE.md sync status, outdated files

## Functional Requirements

- **FR-001:** CLI MUST detect if current directory is a git repo (for project mode)
- **FR-002:** CLI MUST present interactive prompts for setup type and tool selection
- **FR-003:** CLI MUST create the complete `docs/` structure with all AGENTS.md context files
- **FR-004:** CLI MUST create tool-specific directories based on selected tools
- **FR-005:** CLI MUST adapt agent/skill content to each tool's format (frontmatter differences)
- **FR-006:** CLI MUST create root AGENTS.md and CLAUDE.md with placeholders
- **FR-007:** CLI MUST create .githooks/pre-commit for sync verification
- **FR-008:** CLI MUST create CODEOWNERS template
- **FR-009:** CLI MUST NOT overwrite existing files without confirmation
- **FR-010:** CLI SHOULD display clear next-steps after setup

## Constraints

- Node.js 18+ (for npx compatibility)
- Zero runtime dependencies beyond the npm package itself
- Must work on macOS and Linux
- Interactive prompts must have non-interactive fallback (flags for CI/scripts)

## Success Criteria

- SC-1: `npx @org/ai-setup init` produces a working setup in under 30 seconds
- SC-2: Setup passes `npx @org/ai-setup doctor` with zero issues
- SC-3: Two devs on the same team get identical structures from the same selections

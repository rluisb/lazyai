# Research: Post-Install Automation & External Tool Integrations

> **Spec ID:** 003
> **Status:** Research
> **Date:** 2026-04-06
> **Scope:** Post-install automation (knowledge map, extract-standards, placeholder replacement), MCP architecture review, project management tool integrations, documentation tool integrations

---

## 1. Problem Statement

After `ai-setup init` completes, users face several manual steps:

1. **Placeholder hell** — Root files (AGENTS.md, CLAUDE.md, GEMINI.md, copilot-instructions.md) contain 30+ `[YOUR_*]` placeholders that users must fill manually. The `outroSuccess()` function explicitly tells users: *"Open {files} and fill in the [YOUR_*] placeholders"*.

2. **Knowledge Map is a skeleton** — `KNOWLEDGE_MAP.template.md` is scaffolded with placeholder tables (`[specs/features/001-name/]`, `[src/module-a/]`) but never populated with real project data.

3. **Extract-standards is a skill, not automation** — The `/extract-standards` skill exists in `library/skills/extract-standards.md` but is a manual AI-driven process. It's never triggered automatically.

4. **MCP catalog is static** — `library/mcp/catalog.json` has a fixed set of 12 servers. No mechanism exists for users to add project management or documentation tool MCPs during init.

5. **No external tool integration** — Teams using Jira, Linear, Confluence, Notion, etc. have no way to connect those tools through ai-setup.

---

## 2. Current Architecture Analysis

### 2.1 Placeholder System (Two Separate Systems)

**System A: Template Compiler (compiled root files)**
- Uses `{{VARIABLE_NAME}}` syntax via `FragmentResolver`
- Variables resolved at compile time from `FragmentContext`
- Already auto-fills: `PROJECT_NAME`, `PLANNING_DIR`, `PRIMARY_LANGUAGE`, `FRAMEWORK`, `TEST_FRAMEWORK`, `PACKAGE_MANAGER`, `TEST_COMMAND`, `LINT_COMMAND`, `BUILD_COMMAND`, `DEV_COMMAND`, `INSTALL_COMMAND`, `PROJECT_DESCRIPTION`
- Source: `src/compiler/fragment-resolver.ts`
- Used by: `src/scaffold/compiled-root.ts` → writes AGENTS.md, CLAUDE.md, etc.

**System B: Legacy Templates (root/*.template.md)**
- Uses `[YOUR_*]` syntax — NOT resolved by any code
- Files: `AGENTS.template.md`, `CLAUDE.template.md`, `GEMINI.template.md`, `copilot-instructions.template.md`
- These are **not used by the compiled root system** — they appear to be legacy/reference templates
- The compiled root system uses `tool-templates/shared/root.template.md` + fragments instead

**Key Finding:** The `[YOUR_*]` placeholders in `library/root/*.template.md` are NOT the same files that get compiled. The compiled system (`tool-templates/shared/root.template.md`) uses `{{VARIABLE}}` syntax and already resolves many values via `detectProjectStack()`. However, the compiled output still has gaps:
- `{{PROJECT_INSTRUCTIONS}}` resolves to empty string
- `{{FRAMEWORK}}` resolves to empty for unknown project types
- No codebase map is generated
- No conventions are extracted
- No "Do NOT" rules are project-specific

### 2.2 Knowledge Map Generation

- Template: `library/infra/KNOWLEDGE_MAP.template.md`
- Only replacement: `[YOUR_PROJECT_NAME]` → `projectName`
- Scaffolded to: `targetDir/KNOWLEDGE_MAP.md`
- Source: `src/scaffold/infra.ts`
- **Gap:** Could auto-populate "Key Modules" table by scanning `src/` directories, and "Rules & Standards" by listing `specs/rules/` and `specs/standards/` files.

### 2.3 Repo Detection (Already Exists!)

`src/utils/repo-detection.ts` already detects:
- Language (Ruby, TypeScript, Go, Rust, Python)
- Framework (Rails, Next.js, React)
- Package manager (npm, yarn, pnpm, bun, bundle, cargo, pip)
- Test framework (Vitest, Jest, Mocha, go test, cargo test, pytest, RSpec)
- Commands (test, lint, build, dev, install)
- Description (from package.json, Cargo.toml, pyproject.toml)

This data is already passed to `FragmentContext` and used in compiled root files. The gap is that it doesn't go far enough — it doesn't scan directory structure, detect ORM/database, or extract conventions.

### 2.4 MCP Architecture

**Canonical store:** `.ai/mcp.json` — contains ALL servers with `enabled: true/false`
**Compilation:** `src/adapters/mcp-compiler.ts` reads canonical, filters enabled, writes tool-native formats:
- Claude Code / Pi → `.mcp.json` (standard MCP format)
- OpenCode → `opencode.jsonc` (merged into existing config)
- Copilot → `.vscode/mcp.json` (stdio/sse format)
- Gemini → `.gemini/settings.json` (no remote server support)

**Catalog structure per server:**
```json
{
  "description": "...",
  "command": "npx",
  "args": ["-y", "package-name"],
  "env": { "TOKEN": "${TOKEN}" },
  "tools": ["tool1", "tool2"],
  "enabled": true/false,
  "requiresInstall": true/false,
  "installHint": "brew install ..."
}
```

Also supports remote MCP servers via `url` + `headers` (used by context7, atlassian).

**Current catalog (12 servers):**

| Server | Category | Enabled by Default |
|--------|----------|-------------------|
| memory | AI memory | ✅ |
| github | Code hosting | ✅ |
| filesystem | File access | ✅ |
| ripgrep | Code search | ✅ |
| memoria | Git history | ✅ |
| codegraph | Code graph | ❌ (requires install) |
| qmd | Knowledge base | ❌ (requires install) |
| playwright | Browser | ❌ |
| context7 | Library docs | ❌ |
| atlassian | Jira + Confluence | ❌ |
| brave-search | Web search | ❌ |
| fetch | HTTP fetch | ❌ |

### 2.5 Wizard Flow (Where Automation Could Hook In)

Current flow in `src/wizard/index.ts`:
1. Phase 1: Context (scope, tools, name, repos)
2. Phase 2: Features (planning dir, feature flags, git conventions)
3. Phase 3: Conflicts (strategy for existing files)
4. Phase 4: Confirm
5. **Install files** (scaffold specs, constitution, mcp, templates, rules, infra, compiled root, agents/skills/prompts)
6. Write store
7. `outroSuccess()` — tells user to fill placeholders

**Gap:** No Phase 5 "post-install automation" exists. After files are written, nothing runs to enhance them.

---

## 3. External Tool Integration Research

### 3.1 Project Management Tools — MCP Ecosystem

| Platform | Best MCP Server | Install | Official? | Auth | Maturity |
|----------|----------------|---------|-----------|------|----------|
| **Jira** | Atlassian Remote MCP | Remote URL (OAuth) | ✅ Official | OAuth | Stable |
| **Jira** (alt) | `mcp-atlassian` (Python) | `uvx mcp-atlassian` | Community (4.8k⭐) | API token | Stable |
| **Linear** | `mcp-server-linear` | `npx mcp-server-linear` | Community | API key | Stable |
| **Asana** | `@roychri/mcp-server-asana` | `npx @roychri/mcp-server-asana` | Community | PAT | Stable |
| **Monday.com** | `@mondaydotcomorg/monday-api-mcp` | `npx @mondaydotcomorg/monday-api-mcp` | ✅ Official | API key | Stable |
| **Shortcut** | `@shortcut/mcp` | `npx @shortcut/mcp` | ✅ Official | API token | Stable |
| **ClickUp** | `@taazkareem/clickup-mcp-server` | `npx @taazkareem/clickup-mcp-server` | Community | API key | Stable |
| **Notion** | `@notionhq/notion-mcp-server` | `npx @notionhq/notion-mcp-server` | ✅ Official | Integration token | Stable |
| **GitHub** | `github/github-mcp-server` | Docker or Remote | ✅ Official (28.6k⭐) | PAT/OAuth | Stable |
| **GitLab** | `@zereight/mcp-gitlab` | `npx @zereight/mcp-gitlab` | Community | PAT | Stable |
| **Trello** | `mcp-trello` | `npx mcp-trello` | Community | API key + token | Stable |
| **Azure DevOps** | `@tiberriver256/mcp-server-azure-devops` | `npx @tiberriver256/mcp-server-azure-devops` | Community | PAT | Beta |

### 3.2 Documentation Tools — MCP Ecosystem

| Platform | Best MCP Server | Install | Official? | Auth | Maturity |
|----------|----------------|---------|-----------|------|----------|
| **Confluence** | Atlassian Remote MCP | Remote URL (OAuth) | ✅ Official | OAuth | Stable |
| **Confluence** (alt) | `@aashari/mcp-server-atlassian-confluence` | `npx @aashari/...` | Community | API token | Stable |
| **Notion** | `@notionhq/notion-mcp-server` | `npx @notionhq/notion-mcp-server` | ✅ Official | Integration token | Stable |
| **Google Docs** | `@piotr-agier/google-drive-mcp` | `npx @piotr-agier/...` | Community | OAuth | Beta |
| **Coda** | `coda-mcp` | `npx coda-mcp` | Community | API token | Beta |
| **GitBook** | ❌ None exists | — | — | — | — |
| **ReadMe** | ❌ None exists | — | — | — | — |
| **Docusaurus** | ❌ N/A (file-based) | — | — | — | — |
| **Mintlify** | ❌ N/A (file-based) | — | — | — | — |

### 3.3 CLI Tools (Alternative Integration Path)

| Platform | CLI | Install | Scriptable? |
|----------|-----|---------|-------------|
| **Jira** | `jira-cli` (community) | `brew install ankitpokhrel/jira-cli/jira-cli` | ✅ `--plain`/`--raw` output |
| **GitHub** | `gh` (official) | `brew install gh` | ✅ JSON output, GraphQL |
| **GitLab** | `glab` (official) | `brew install glab` | ✅ JSON output |
| **Azure DevOps** | `az devops` (official) | `az extension add --name azure-devops` | ✅ `--output json` |
| **ReadMe** | `rdme` (official) | `npm install -g rdme` | ✅ Non-interactive modes |
| **Linear** | SDK only | `npm install @linear/sdk` | ✅ TypeScript SDK |
| **Notion** | SDK only | `npm install @notionhq/client` | ✅ TypeScript SDK |

### 3.4 Integration Strategy Matrix

For each tool, there are up to 3 integration paths:

| Path | Mechanism | When to Use |
|------|-----------|-------------|
| **MCP Server** | Add to `.ai/mcp.json` catalog, compile to tool-native format | AI agent needs real-time access to the tool during coding sessions |
| **Skill** | Add to `.ai/skills/` as a workflow prompt | AI agent needs a structured workflow (e.g., "sync issue status", "update docs") |
| **CLI wrapper** | Add to `cliTools` in catalog or as a shell command in skills | Tool has a CLI that can be called non-interactively |

**Recommendation:** MCP is the primary path. Skills complement MCPs for structured workflows. CLI is fallback when no MCP exists.

---

## 4. Proposed Solutions

### 4.1 Post-Install Automation: Auto-Fill Placeholders

**What:** After scaffolding, run `detectProjectStack()` results through the compiled root template system to fill as many `{{VARIABLE}}` values as possible. For the legacy `[YOUR_*]` templates, either:
- (a) Migrate them to use `{{VARIABLE}}` syntax and compile them, or
- (b) Deprecate them since the compiled root system already supersedes them

**Current state:** The compiled root system (`tool-templates/shared/root.template.md`) already resolves most variables. The remaining gaps are:
- `{{PROJECT_INSTRUCTIONS}}` — empty
- `{{PROJECT_DESCRIPTION}}` — only filled if package.json has `description`
- No codebase map section in the compiled template
- No conventions section

**Proposed enhancement — "Smart Context Builder":**
1. Scan directory structure → generate codebase map table
2. Scan for database files (prisma, drizzle, typeorm, sequelize, activerecord) → detect ORM/database
3. Scan for `.env` files → detect required environment variables (names only, not values)
4. Scan for CI/CD files → detect pipeline tool
5. Scan for Docker files → detect containerization
6. Generate a `PROJECT_INSTRUCTIONS` block from all detected context

**Effort:** Medium — extends existing `detectProjectStack()` with deeper scanning.

### 4.2 Post-Install Automation: Knowledge Map Builder

**What:** After scaffolding, auto-populate `KNOWLEDGE_MAP.md` with real data.

**Proposed approach:**
1. Scan `src/` (or equivalent) top-level directories → populate "Key Modules" table
2. List `specs/rules/*.md` files → populate "Rules & Standards" table
3. List `specs/standards/*.md` files → add to same table
4. If `specs/adrs/` has files → populate "Architecture Decisions" table
5. If `specs/features/` has directories → populate "Active Features" table

**Effort:** Low-Medium — mostly file listing and template interpolation.

### 4.3 Post-Install Automation: Extract Standards Trigger

**What:** After init, optionally run a lightweight version of `/extract-standards` that:
1. Checks if the codebase has enough code (YAGNI gate from the skill)
2. If yes, suggests running `/extract-standards` as a next step
3. If the user opts in (interactive mode), runs the extraction

**Proposed approach:**
- Add a Phase 5 to the wizard: "Post-Install Intelligence"
- In interactive mode: ask "Would you like to scan your codebase for patterns?"
- In non-interactive mode: skip (or add `--extract-standards` flag)
- The actual extraction still uses the AI skill — we just trigger it

**Effort:** Low — mostly UX flow, the skill already exists.

### 4.4 MCP Catalog Expansion: External Tool Integrations

**What:** Expand `library/mcp/catalog.json` with project management and documentation tool servers. Add wizard step to select which external tools to enable.

**Proposed catalog additions (Tier 1 — stable, well-maintained):**

```json
{
  "linear": {
    "description": "Linear project management",
    "command": "npx",
    "args": ["-y", "mcp-server-linear"],
    "env": { "LINEAR_API_KEY": "${LINEAR_API_KEY}" },
    "tools": ["list_issues", "create_issue", "update_issue", "list_projects"],
    "enabled": false,
    "requiresInstall": false,
    "category": "project-management"
  },
  "asana": {
    "description": "Asana task and project management",
    "command": "npx",
    "args": ["-y", "@roychri/mcp-server-asana"],
    "env": { "ASANA_PAT": "${ASANA_PAT}" },
    "tools": ["search_tasks", "get_task", "create_task", "update_task"],
    "enabled": false,
    "requiresInstall": false,
    "category": "project-management"
  },
  "monday": {
    "description": "Monday.com board and item management",
    "command": "npx",
    "args": ["-y", "@mondaydotcomorg/monday-api-mcp"],
    "env": { "MONDAY_API_KEY": "${MONDAY_API_KEY}" },
    "tools": ["list_boards", "get_items", "create_item", "update_item"],
    "enabled": false,
    "requiresInstall": false,
    "category": "project-management"
  },
  "shortcut": {
    "description": "Shortcut story and project management",
    "command": "npx",
    "args": ["-y", "@shortcut/mcp"],
    "env": { "SHORTCUT_API_TOKEN": "${SHORTCUT_API_TOKEN}" },
    "tools": ["search_stories", "get_story", "create_story", "list_epics"],
    "enabled": false,
    "requiresInstall": false,
    "category": "project-management"
  },
  "clickup": {
    "description": "ClickUp task and document management",
    "command": "npx",
    "args": ["-y", "@taazkareem/clickup-mcp-server"],
    "env": { "CLICKUP_API_KEY": "${CLICKUP_API_KEY}" },
    "tools": ["list_tasks", "create_task", "update_task", "list_spaces"],
    "enabled": false,
    "requiresInstall": false,
    "category": "project-management"
  },
  "notion": {
    "description": "Notion pages, databases, and workspace management",
    "command": "npx",
    "args": ["-y", "@notionhq/notion-mcp-server"],
    "env": { "OPENAPI_MCP_HEADERS": "{\"Authorization\": \"Bearer ${NOTION_TOKEN}\", \"Notion-Version\": \"2022-06-28\"}" },
    "tools": ["search", "get_page", "create_page", "query_database"],
    "enabled": false,
    "requiresInstall": false,
    "category": "project-management"
  },
  "gitlab": {
    "description": "GitLab repository, issue, and MR management",
    "command": "npx",
    "args": ["-y", "@zereight/mcp-gitlab"],
    "env": { "GITLAB_TOKEN": "${GITLAB_TOKEN}", "GITLAB_URL": "${GITLAB_URL}" },
    "tools": ["list_issues", "create_issue", "list_merge_requests"],
    "enabled": false,
    "requiresInstall": false,
    "category": "project-management"
  },
  "trello": {
    "description": "Trello board and card management",
    "command": "npx",
    "args": ["-y", "mcp-trello"],
    "env": { "TRELLO_API_KEY": "${TRELLO_API_KEY}", "TRELLO_TOKEN": "${TRELLO_TOKEN}" },
    "tools": ["list_boards", "get_cards", "create_card", "update_card"],
    "enabled": false,
    "requiresInstall": false,
    "category": "project-management"
  },
  "azure-devops": {
    "description": "Azure DevOps work items, repos, and pipelines",
    "command": "npx",
    "args": ["-y", "@tiberriver256/mcp-server-azure-devops"],
    "env": { "AZURE_DEVOPS_ORG_URL": "${AZURE_DEVOPS_ORG_URL}", "AZURE_DEVOPS_PAT": "${AZURE_DEVOPS_PAT}" },
    "tools": ["list_work_items", "create_work_item", "list_repositories"],
    "enabled": false,
    "requiresInstall": false,
    "category": "project-management"
  },
  "confluence": {
    "description": "Confluence page and space management",
    "command": "npx",
    "args": ["-y", "@aashari/mcp-server-atlassian-confluence"],
    "env": {
      "CONFLUENCE_URL": "${CONFLUENCE_URL}",
      "CONFLUENCE_USERNAME": "${CONFLUENCE_USERNAME}",
      "CONFLUENCE_API_TOKEN": "${CONFLUENCE_API_TOKEN}"
    },
    "tools": ["search_pages", "get_page", "create_page", "list_spaces"],
    "enabled": false,
    "requiresInstall": false,
    "category": "documentation"
  },
  "google-docs": {
    "description": "Google Docs, Sheets, and Drive access",
    "command": "npx",
    "args": ["-y", "@piotr-agier/google-drive-mcp"],
    "tools": ["list_files", "get_document", "create_document"],
    "enabled": false,
    "requiresInstall": false,
    "category": "documentation"
  }
}
```

**Wizard integration:**
- Add a new wizard step (Phase 1.5 or Phase 2 extension): "External Tools"
- Group by category: Project Management, Documentation, Code Hosting
- Show only relevant options (e.g., don't show GitHub MCP if `gh` CLI is already detected)
- For each selected tool, note required env vars for the user

### 4.5 MCP Category System

**What:** Add a `category` field to catalog entries for better organization and wizard UX.

**Proposed categories:**

| Category | Description | Examples |
|----------|-------------|---------|
| `core` | Essential for AI coding | memory, filesystem, ripgrep |
| `code-hosting` | Repository management | github, gitlab |
| `code-intelligence` | Code analysis | codegraph, memoria, context7 |
| `project-management` | Issue/task tracking | jira/atlassian, linear, asana, monday, shortcut, clickup, notion, trello, azure-devops |
| `documentation` | Docs platforms | confluence, notion, google-docs, coda |
| `browser` | Web interaction | playwright, fetch, brave-search |

### 4.6 Skill Templates for External Tools

**What:** Create skill templates that complement MCP servers with structured workflows.

**Proposed skills:**
1. **`sync-issues`** — Sync issue status between external PM tool and specs/features/
2. **`update-docs`** — Push specs/standards/ content to external docs platform
3. **`import-requirements`** — Pull requirements from PM tool into specs/features/ research.md

These would be parameterized by the selected PM/docs tool.

---

## 5. Implementation Waves

### Wave 1: Post-Install Intelligence (Low-Medium effort)
1. **Enhanced `detectProjectStack()`** — Add ORM, database, CI/CD, Docker detection
2. **Smart Knowledge Map builder** — Auto-populate from directory scan
3. **Codebase map generator** — Scan top-level src/ dirs for compiled root
4. **Phase 5 wizard step** — "Post-install: would you like to scan for patterns?"

### Wave 2: MCP Catalog Expansion (Medium effort)
1. **Add category field** to catalog schema
2. **Add 11 new servers** to catalog (PM + docs tools)
3. **Wizard step for external tools** — grouped selection by category
4. **Env var guidance** — after init, list required env vars for enabled servers

### Wave 3: Skill Templates for Integrations (Medium effort)
1. **`sync-issues` skill** — parameterized by PM tool
2. **`update-docs` skill** — parameterized by docs tool
3. **`import-requirements` skill** — pull from PM into specs/

### Wave 4: Deep Automation (Higher effort)
1. **Auto-extract conventions** — lightweight version of extract-standards during init
2. **Auto-generate PROJECT_INSTRUCTIONS** — from all detected context
3. **Deprecate legacy `[YOUR_*]` templates** — or migrate to compiled system

---

## 6. Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| MCP servers break/change APIs | Users get broken integrations | Pin versions in catalog, add health check in `doctor` command |
| Too many wizard steps | UX fatigue | Group external tools into single multi-select, skip if `--no-interactive` |
| Auto-detection wrong | Incorrect project context | Always show detected values for confirmation in interactive mode |
| Env var confusion | Users don't know what tokens to set | Generate `.env.example` with required vars, add setup instructions |
| Category proliferation | Catalog becomes unwieldy | Limit to 6 categories, use presets (e.g., "standard" doesn't include PM tools) |

---

## 7. Open Questions

1. **Should PM/docs MCPs be part of presets?** Or always opt-in regardless of preset level?
2. **Should we support the Atlassian Remote MCP (OAuth)?** It's official but requires OAuth flow, not just env vars.
3. **Should the Knowledge Map builder run automatically or be opt-in?** Auto-running is convenient but might generate noise for small projects.
4. **Should we generate `.env.example`?** Listing required env vars for enabled MCP servers.
5. **How deep should auto-detection go?** Just top-level dirs, or recursive scanning for module boundaries?
6. **Should legacy `library/root/*.template.md` files be removed?** They're not used by the compiled system but exist in the library.

---

## 8. Appendix: Full MCP Server Research Data

### Project Management — All Known MCP Servers

| Platform | Package | Official | Stars | Maturity |
|----------|---------|----------|-------|----------|
| Jira | Atlassian Remote MCP | ✅ | N/A | Stable |
| Jira | `mcp-atlassian` (Python) | ❌ | 4.8k | Stable |
| Jira | `mcp-jira-cloud` | ❌ | — | Stable |
| Jira | `@aashari/mcp-server-atlassian-jira` | ❌ | — | Stable |
| Jira | `mcp-jira-confluence` | ❌ | — | Beta |
| Linear | `mcp-server-linear` | ❌ | — | Stable |
| Linear | `@tacticlaunch/mcp-linear` | ❌ | — | Stable |
| Asana | `@roychri/mcp-server-asana` | ❌ | — | Stable |
| Asana | `@panchopoliti/mcp-server-asana` | ❌ | — | Beta |
| Monday.com | `@mondaydotcomorg/monday-api-mcp` | ✅ | — | Stable |
| Shortcut | `@shortcut/mcp` | ✅ | — | Stable |
| ClickUp | `@taazkareem/clickup-mcp-server` | ❌ | — | Stable |
| ClickUp | `clickup-mcp-server` | ❌ | — | Stable |
| Notion | `@notionhq/notion-mcp-server` | ✅ | — | Stable |
| Notion | `@suekou/mcp-notion-server` | ❌ | — | Stable |
| GitHub | `github/github-mcp-server` | ✅ | 28.6k | Stable |
| GitHub | `mcp-github-project-manager` | ❌ | — | Beta |
| GitLab | `@zereight/mcp-gitlab` | ❌ | — | Stable |
| GitLab | `@structured-world/gitlab-mcp` | ❌ | — | Stable |
| Trello | `mcp-trello` | ❌ | — | Stable |
| Azure DevOps | `@tiberriver256/mcp-server-azure-devops` | ❌ | — | Beta |
| Azure DevOps | `azure-devops-mcp` (read-only) | ❌ | — | Beta |

### Documentation — All Known MCP Servers

| Platform | Package | Official | Maturity |
|----------|---------|----------|----------|
| Confluence | Atlassian Remote MCP | ✅ | Stable |
| Confluence | `@aashari/mcp-server-atlassian-confluence` | ❌ | Stable |
| Confluence | `@dsazz/mcp-confluence` | ❌ | Beta |
| Notion | `@notionhq/notion-mcp-server` | ✅ | Stable |
| Google Docs | `@piotr-agier/google-drive-mcp` | ❌ | Beta |
| Google Docs | `@a-bonus/google-docs-mcp` | ❌ | Beta |
| Coda | `coda-mcp` | ❌ | Beta |
| GitBook | ❌ None exists | — | — |
| ReadMe | ❌ None exists | — | — |
| Docusaurus | ❌ N/A (file-based) | — | — |
| Mintlify | ❌ N/A (file-based) | — | — |
| Swimm | ❌ None exists | — | — |
| Slite | ❌ None exists | — | — |

### CLI Tools — Alternative Integration Path

| Platform | CLI Tool | Install | Scriptable |
|----------|----------|---------|------------|
| Jira | `jira-cli` (community, 5.5k⭐) | `brew install ankitpokhrel/jira-cli/jira-cli` | ✅ `--plain`/`--raw` |
| GitHub | `gh` (official) | `brew install gh` | ✅ JSON, GraphQL |
| GitLab | `glab` (official) | `brew install glab` | ✅ JSON output |
| Azure DevOps | `az devops` (official) | `az extension add --name azure-devops` | ✅ `--output json` |
| ReadMe | `rdme` (official) | `npm install -g rdme` | ✅ Non-interactive |
| Linear | `@linear/sdk` (SDK) | `npm install @linear/sdk` | ✅ TypeScript |
| Notion | `@notionhq/client` (SDK) | `npm install @notionhq/client` | ✅ TypeScript |

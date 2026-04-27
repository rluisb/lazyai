# Graph Report - .  (2026-04-27)

## Corpus Check
- Large corpus: 724 files · ~943,825 words. Semantic extraction will be expensive (many Claude tokens). Consider running on a subfolder, or use --no-semantic to run AST-only.

## Summary
- 3529 nodes · 8426 edges · 96 communities detected
- Extraction: 61% EXTRACTED · 39% INFERRED · 0% AMBIGUOUS · INFERRED: 3258 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Adapter Scope Testing|Adapter Scope Testing]]
- [[_COMMUNITY_Tool Adapters|Tool Adapters]]
- [[_COMMUNITY_Catalog & Resolver|Catalog & Resolver]]
- [[_COMMUNITY_CLI Add & Codex Commands|CLI Add & Codex Commands]]
- [[_COMMUNITY_Copilot Headless Runner|Copilot Headless Runner]]
- [[_COMMUNITY_CLI Import & Migrate|CLI Import & Migrate]]
- [[_COMMUNITY_Adapter Registry|Adapter Registry]]
- [[_COMMUNITY_Claude & OpenCode CLI|Claude & OpenCode CLI]]
- [[_COMMUNITY_CLI Catalog & Eject|CLI Catalog & Eject]]
- [[_COMMUNITY_Wizard Phases|Wizard Phases]]
- [[_COMMUNITY_Library Agents & Skills|Library Agents & Skills]]
- [[_COMMUNITY_Spec Planning & Research|Spec Planning & Research]]
- [[_COMMUNITY_MCP & Configuration|MCP & Configuration]]
- [[_COMMUNITY_Agent Teams & Orchestration|Agent Teams & Orchestration]]
- [[_COMMUNITY_Security & Rules|Security & Rules]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 34|Community 34]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Community 37|Community 37]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 43|Community 43]]
- [[_COMMUNITY_Community 44|Community 44]]
- [[_COMMUNITY_Community 45|Community 45]]
- [[_COMMUNITY_Community 46|Community 46]]
- [[_COMMUNITY_Community 47|Community 47]]
- [[_COMMUNITY_Community 48|Community 48]]
- [[_COMMUNITY_Community 49|Community 49]]
- [[_COMMUNITY_Community 50|Community 50]]
- [[_COMMUNITY_Community 51|Community 51]]
- [[_COMMUNITY_Community 52|Community 52]]
- [[_COMMUNITY_Community 53|Community 53]]
- [[_COMMUNITY_Community 54|Community 54]]
- [[_COMMUNITY_Community 55|Community 55]]
- [[_COMMUNITY_Community 58|Community 58]]
- [[_COMMUNITY_Community 60|Community 60]]
- [[_COMMUNITY_Community 61|Community 61]]
- [[_COMMUNITY_Community 62|Community 62]]
- [[_COMMUNITY_Community 63|Community 63]]
- [[_COMMUNITY_Community 64|Community 64]]
- [[_COMMUNITY_Community 65|Community 65]]
- [[_COMMUNITY_Community 66|Community 66]]
- [[_COMMUNITY_Community 67|Community 67]]
- [[_COMMUNITY_Community 68|Community 68]]
- [[_COMMUNITY_Community 69|Community 69]]
- [[_COMMUNITY_Community 70|Community 70]]
- [[_COMMUNITY_Community 71|Community 71]]
- [[_COMMUNITY_Community 72|Community 72]]
- [[_COMMUNITY_Community 73|Community 73]]
- [[_COMMUNITY_Community 74|Community 74]]
- [[_COMMUNITY_Community 75|Community 75]]
- [[_COMMUNITY_Community 85|Community 85]]
- [[_COMMUNITY_Community 86|Community 86]]
- [[_COMMUNITY_Community 87|Community 87]]
- [[_COMMUNITY_Community 88|Community 88]]
- [[_COMMUNITY_Community 89|Community 89]]
- [[_COMMUNITY_Community 90|Community 90]]
- [[_COMMUNITY_Community 91|Community 91]]
- [[_COMMUNITY_Community 92|Community 92]]
- [[_COMMUNITY_Community 93|Community 93]]
- [[_COMMUNITY_Community 94|Community 94]]
- [[_COMMUNITY_Community 95|Community 95]]
- [[_COMMUNITY_Community 96|Community 96]]
- [[_COMMUNITY_Community 97|Community 97]]
- [[_COMMUNITY_Community 98|Community 98]]
- [[_COMMUNITY_Community 99|Community 99]]
- [[_COMMUNITY_Community 100|Community 100]]
- [[_COMMUNITY_Community 101|Community 101]]
- [[_COMMUNITY_Community 102|Community 102]]
- [[_COMMUNITY_Community 103|Community 103]]
- [[_COMMUNITY_Community 104|Community 104]]
- [[_COMMUNITY_Community 105|Community 105]]
- [[_COMMUNITY_Community 106|Community 106]]
- [[_COMMUNITY_Community 107|Community 107]]
- [[_COMMUNITY_Community 108|Community 108]]

## God Nodes (most connected - your core abstractions)
1. `readFile()` - 145 edges
2. `fileExists()` - 142 edges
3. `writeFile()` - 132 edges
4. `run()` - 91 edges
5. `ensureDir()` - 91 edges
6. `fileHash()` - 69 edges
7. `contains()` - 41 edges
8. `Scan()` - 40 edges
9. `runWizard()` - 40 edges
10. `DirExists()` - 35 edges

## Surprising Connections (you probably didn't know these)
- `Agent Harness (CLAUDE.md)` --semantically_similar_to--> `Agent Harness (AGENTS.md)`  [INFERRED] [semantically similar]
  CLAUDE.md → AGENTS.md
- `main()` --calls--> `Execute()`  [INFERRED]
  main.go → internal/migration/executor.go
- `TestRunPhase1NonInteractiveDefaults()` --calls--> `runPhase1()`  [INFERRED]
  tui/wizard/phase1_test.go → src/wizard/phase1-context.ts
- `TestResolveToolRoot_AllPairs()` --calls--> `resolveToolRoot()`  [INFERRED]
  internal/adapter/scope_test.go → src/adapters/mcp-compiler.ts
- `TestResolveToolRoot_NilCtx()` --calls--> `resolveToolRoot()`  [INFERRED]
  internal/adapter/scope_test.go → src/adapters/mcp-compiler.ts

## Hyperedges (group relationships)
- **Canonical Source to Compile Workflow** — readme_ai_setup, readme_canonical_source_model, readme_ai_dir, readme_compile_output, readme_ai_setup_json [EXTRACTED 1.00]
- **Supported AI Tools** — readme_opencode, readme_claude_code, readme_gemini_cli, readme_github_copilot, readme_codex [EXTRACTED 1.00]
- **Project Scope Supported Tools** — readme_project_scope, readme_opencode, readme_claude_code, readme_gemini_cli, readme_github_copilot, readme_codex [EXTRACTED 1.00]
- **Global Scope Supported Tools** — readme_global_scope, readme_opencode, readme_claude_code [EXTRACTED 1.00]
- **Bundled Agents** — readme_bundled_agents, readme_agent_builder, readme_agent_documenter, readme_agent_orchestrator, readme_agent_planner, readme_agent_red_team, readme_agent_reviewer, readme_agent_scout [EXTRACTED 1.00]
- **Bundled Skills** — readme_bundled_skills, readme_skill_anti_speculation, readme_skill_extract_standards, readme_skill_implement, readme_skill_iterate, readme_skill_memory_write, readme_skill_parallel_execution, readme_skill_plan, readme_skill_research, readme_skill_tdd_loop [EXTRACTED 1.00]
- **Bundled Orchestration Definitions** — readme_bundled_orchestration, readme_chain_feature, readme_chain_bugfix, readme_chain_review, readme_chain_refactor, readme_chain_tdd, readme_chain_onboard, readme_team_feature_team, readme_team_review_team, readme_team_assessment_team, readme_domain_backend, readme_domain_frontend, readme_domain_data, readme_domain_devops, readme_domain_security, readme_mode_autonomous, readme_mode_junior, readme_mode_senior [EXTRACTED 1.00]
- **CLAUDE.md Workflow System** — claude_rpi_workflow, claude_reasoning_protocol, claude_decision_protocol, claude_quality_gates, claude_agent_harness, claude_bug_resolution, claude_git_conventions [INFERRED 0.80]
- **AGENTS.md Workflow System** — agents_rpi_workflow, agents_reasoning_protocol, agents_decision_protocol, agents_quality_gates, agents_agent_harness, agents_bug_resolution, agents_git_conventions [INFERRED 0.80]
- **Cross-File Workflow Alignment** — claude_rpi_workflow, agents_rpi_workflow, claude_reasoning_protocol, agents_reasoning_protocol, claude_decision_protocol, agents_decision_protocol, claude_quality_gates, agents_quality_gates, claude_agent_harness, agents_agent_harness, claude_bug_resolution, agents_bug_resolution, claude_git_conventions, agents_git_conventions [INFERRED 0.85]

## Communities

### Community 0 - "Adapter Scope Testing"
Cohesion: 0.02
Nodes (309): newScopeTestContext(), TestAdapter_ScopeParity(), TestCodexAdapter_CompileMCP_WritesServers(), TestCodexAdapter_ConfigMergePreservesUserKeys(), TestCodexAdapter_WritesConfigAndSplitSkills(), TestCopilotAdapter_GlobalScope_Emits(), TestCopilotAdapter_GlobalScope_Skips(), TestGeminiAdapter_DriveCLI_CallsGeminiBinary() (+301 more)

### Community 1 - "Tool Adapters"
Cohesion: 0.01
Nodes (180): ClaudeCodeAdapter, GeminiAdapter, OpenCodeAdapter, PiAdapter, AdapterRegistry, getSelectionSet(), registerAdd(), parseTools() (+172 more)

### Community 2 - "Catalog & Resolver"
Cohesion: 0.02
Nodes (189): importAgent(), importSkill(), dbRowToAgent(), dbRowToSkill(), resolveCatalog(), validateFrontmatterForKind(), parseConnectArgs(), runConnect() (+181 more)

### Community 3 - "CLI Add & Codex Commands"
Cohesion: 0.02
Nodes (172): CodexAdapter, mergeAgentIds(), mergeSkillIds(), mergeStringSlices(), mergeToolIds(), runAdd(), runAddInteractive(), runAddNonInteractive() (+164 more)

### Community 4 - "Copilot Headless Runner"
Cohesion: 0.03
Nodes (169): CopilotCLIRunner, DefaultCopilotCLIRunner, TestCanRunHeadless_Values(), TestRunHeadlessValidation_ClaudeCode_Installed(), TestRunHeadlessValidation_Codex_Installed(), TestRunHeadlessValidation_NoOpAdapters(), TestRunHeadlessValidation_ToolNotInstalled(), TestRunHeadlessValidation_UsesTargetDir() (+161 more)

### Community 5 - "CLI Import & Migrate"
Cohesion: 0.02
Nodes (117): init(), printImportResult(), runImport(), init(), runMigrate(), Execute(), formatAdapterList(), normalizeStrategy() (+109 more)

### Community 6 - "Adapter Registry"
Cohesion: 0.02
Nodes (59): Registry, NewRegistry(), countCreated(), init(), runCreate(), AgentGenerator, ChainGenerator, toFunctionName() (+51 more)

### Community 7 - "Claude & OpenCode CLI"
Cohesion: 0.03
Nodes (105): LookupClaudeBinary(), TestDefaultClaudeCLIRunner_VersionCommand(), TestLookupClaudeBinary_Found(), ClaudeCLIRunner, DefaultClaudeCLIRunner, installOpenCodePlugins(), makeFakeBin(), TestInstallOpenCodePlugins_BinaryAbsent() (+97 more)

### Community 8 - "CLI Catalog & Eject"
Cohesion: 0.03
Nodes (91): artifactGroup, artifactItem, Catalog, CatalogServer, CheckResult, CheckStatus, runEject(), fileBasename() (+83 more)

### Community 9 - "Wizard Phases"
Cohesion: 0.03
Nodes (96): indentLines(), skillToAgentYAML(), toDisplayName(), CopyLibraryDirectoryOption, InstallToolContextFilesOption, BuildOpenCodeAgentFrontmatter(), inheritedDescription(), sortedStringKeys() (+88 more)

### Community 10 - "Library Agents & Skills"
Cohesion: 0.03
Nodes (97): IsScopeSupported(), projectSubdir(), ResolveCodexRoots(), resolveHomeDir(), ResolveToolRoot(), TestIsScopeSupported(), TestResolveCodexRoots(), TestResolveToolRoot_AllPairs() (+89 more)

### Community 11 - "Spec Planning & Research"
Cohesion: 0.04
Nodes (96): TestToCopilotCLIMcp_EnvPreserved(), TestToCopilotCLIMcp_SseServer(), TestToCopilotCLIMcp_StdioServer(), claudeCliAvailable(), codexSectionPrefix(), compileMcp(), copilotProbePasses(), getEnabledServers() (+88 more)

### Community 12 - "MCP & Configuration"
Cohesion: 0.03
Nodes (88): Agent Harness (AGENTS.md), Bug Resolution (AGENTS.md), Context Discipline, Decision Protocol (AGENTS.md), Git Conventions (AGENTS.md), OpenCode-Specific Notes, Quality Gates (AGENTS.md), Reasoning Protocol (AGENTS.md) (+80 more)

### Community 13 - "Agent Teams & Orchestration"
Cohesion: 0.04
Nodes (79): Permanent Decisions, ADR Rules, Configuration Tampering, Context Poisoning, Privilege Escalation, Prompt Injection, Red-Team Agent, Secret Exposure (+71 more)

### Community 14 - "Security & Rules"
Cohesion: 0.05
Nodes (42): importFromLibrary(), importJsonDefinitions(), CatalogStore, pickFlag(), pickFlagInt(), pickFlags(), runCatalog(), catalogChecksum() (+34 more)

### Community 15 - "Community 15"
Cohesion: 0.06
Nodes (62): Access Rule, Bounded Access Rationale, Adversarial Protection Rationale, Agent Security Rule, ai-setup doctor, Privilege Escalation Threat, Prompt Injection Threat, Secret Exposure Threat (+54 more)

### Community 16 - "Community 16"
Cohesion: 0.07
Nodes (49): CatalogDB, CurrentState, DesiredPath, DesiredRoot, DesiredState, DesiredTarget, Inventory, MCPEntry (+41 more)

### Community 17 - "Community 17"
Cohesion: 0.05
Nodes (53): Access Rule, Path Access Controls, Scope Boundaries, Commit Message Pattern, 5-Why Analysis, OpenCode AGENTS.md Sample, Pre-Flight Task Framing, Plan Prompt (+45 more)

### Community 18 - "Community 18"
Cohesion: 0.07
Nodes (47): init(), main(), AgentsDir(), CodexAssetsDir(), ConstitutionDir(), CopilotAgentsDir(), CopilotInstructionsDir(), FindLibraryDir() (+39 more)

### Community 19 - "Community 19"
Cohesion: 0.08
Nodes (24): contains(), containsStr(), TestCompileForTools(), TestFragmentResolver_Conditional(), TestFragmentResolver_DefaultVariable(), TestFragmentResolver_IncludeFromDisk(), TestFragmentResolver_IncludeFromFS(), TestFragmentResolver_IncludeNotFound() (+16 more)

### Community 20 - "Community 20"
Cohesion: 0.09
Nodes (36): cachedScan(), clearScanCache(), dirMtime(), extractConstraints(), extractFirstParagraph(), mergeScannedCatalogs(), normalizeName(), readAgentsMd() (+28 more)

### Community 21 - "Community 21"
Cohesion: 0.07
Nodes (45): Agents Directory, Agent Progression Levels, Builder Agent, Follow Plan Exactly, No Unrequested Features, Sonnet Model for Builder, Constraint Techniques, Hard Constraints (+37 more)

### Community 22 - "Community 22"
Cohesion: 0.09
Nodes (37): init(), readGeminiExtensionMcpCatalog(), runBuildGeminiExtension(), TestBuildGeminiExtension_AbsentCatalogReturnsNil(), TestBuildGeminiExtension_EmptyCatalogReturnsNil(), TestBuildGeminiExtension_GoldenOutput(), TestBuildGeminiExtension_InvalidJsonReturnsNil(), TestBuildGeminiExtension_ReadsMcpCatalogFromCwd() (+29 more)

### Community 23 - "Community 23"
Cohesion: 0.09
Nodes (37): RemoveAll(), adoptResources(), aiSetupHome(), applyDetectionState(), applyMCPEntryStates(), canImportState(), compareObservedPaths(), copyPath() (+29 more)

### Community 24 - "Community 24"
Cohesion: 0.1
Nodes (23): TestMCPAddJsonPayloadGlobalScope(), TestMCPAddJsonPayloadProjectScope(), TestMCPAddJsonPayloadWorkspaceScope(), TestMCPDisabledServerSkipped(), TestMCPFallbackWhenBinaryMissing(), TestMCPMultipleServersAddedInSequence(), TestMCPPreCheckAndAddJson(), TestMcpServerToJSON_HTTP() (+15 more)

### Community 25 - "Community 25"
Cohesion: 0.09
Nodes (32): Requirements Checklist — 011: OpenCode Deep Setup, Agent Coordination (Scout/Planner/Builder/Reviewer/Documenter), AGENTS.md Context File, OpenCode Commands and Modes Library Assets, Context Discipline for AI Agents, Decision Protocol for Architecture Choices, OpenCode Agent Frontmatter Emitter, MCP Per-Server Deep Merge Strategy (+24 more)

### Community 26 - "Community 26"
Cohesion: 0.08
Nodes (27): Copilot global MCP compile, toCopilotCLIMcp, toCopilotVSCodeMcp, Copilot drive-cli deferred item, ~/.copilot/mcp-config.json, Fang CLI styling, Integration testing end-to-end, Bubble Tea v2 (+19 more)

### Community 27 - "Community 27"
Cohesion: 0.12
Nodes (27): ADR Template, Template Rules and Inventory, Bugfix RCA Template, Checklist Template, Commit Message Pattern, Compozy Adaptation Analysis, Code Review Template, Tech Debt Template (+19 more)

### Community 28 - "Community 28"
Cohesion: 0.12
Nodes (27): AGENTS.md Template, CLAUDE.md Template, Code Review Workflow Rules, AI Compliance and Audit Trail, AI Agentic Setup Implementation Plan, RPI Pattern, Orchestration Usage, Orchestrator Blueprint (+19 more)

### Community 29 - "Community 29"
Cohesion: 0.1
Nodes (25): Global Agents Subdir Fix, Task 001: Global Agents Subdir and Context, Orchestrator Global Scope Bug, Task 002: Orchestrator at Global Scope, Hybrid Provider-Command Strategy, Spec 005: Setup Flow Fixes, Claude CLI Runner Interface, Task 008: Claude CLI Probe and Runner (+17 more)

### Community 30 - "Community 30"
Cohesion: 0.13
Nodes (15): codexExecValidationArgs(), codexMcpServerCount(), countCodexMcpPlaintext(), displayCodexInstallSummary(), parseCodexMcpListJSON(), newCodexTemplateFSWithSkills(), TestCodexAdapter_Install_WritesLibraryAgentsOverride(), TestCodexAgentsOverrideTemplate_ConstantPointsToExistingFile() (+7 more)

### Community 31 - "Community 31"
Cohesion: 0.16
Nodes (17): computeLineDiff(), computeWordDiff(), findHunks(), formatLineNum(), getRequired(), renderDiffPreview(), renderSimpleDiff(), cancelAndExit() (+9 more)

### Community 32 - "Community 32"
Cohesion: 0.14
Nodes (14): Environment Variables, Parameterized Queries, PII Protection, Rate Limiting, Authentication & Authorization, Dependency Security, Security Escalation Triggers, HTTPS & Transport (+6 more)

### Community 33 - "Community 33"
Cohesion: 0.19
Nodes (10): FeatureFlagsForTemplate, FragmentContext, commandsForLanguage(), fallbackMarker(), mustContain(), TestFillClaudeMdPlaceholders(), fillClaudeMdPlaceholders(), memoryDocDestPath() (+2 more)

### Community 34 - "Community 34"
Cohesion: 0.2
Nodes (3): failedEntry, randomAlphaNum(), OperationTracker

### Community 35 - "Community 35"
Cohesion: 0.23
Nodes (12): Simplification Decisions, Extract Standards Skill, Ledger Pattern, Simplification Plan, Preset System, Requirements Checklist, Simplification Research, Scope-Aware Templates (+4 more)

### Community 36 - "Community 36"
Cohesion: 0.22
Nodes (2): trimExt(), PiAdapter

### Community 37 - "Community 37"
Cohesion: 0.25
Nodes (9): Config Merge Helper, CLI Tool Parity Plan, CLI Tool Parity Research, Scope Resolver, Codex DriveCLI, OpenCode Deep Setup, Go/TS Parity Design, Go/TS Parity Plan (+1 more)

### Community 38 - "Community 38"
Cohesion: 0.29
Nodes (2): GlobalConfigDir(), GlobalSetupDir()

### Community 39 - "Community 39"
Cohesion: 0.25
Nodes (8): Bugfix Prompt, Feature Prompt, Refactor Prompt, RPI Prompt, Template Standardization Rationale, Tech Debt Prompt, Template Inventory, Templates Directory

### Community 41 - "Community 41"
Cohesion: 0.33
Nodes (6): better-sqlite3 dependency, Phase 0 scaffolding, Phase 2 SQLite-backed persistence, Phase 3 catalog DB versioning, Persistence seam in persistence.ts, Pure state machines

### Community 42 - "Community 42"
Cohesion: 0.33
Nodes (6): Approval Gate, Bootstrap Workflow, Context Loading Order, Codegraph Integration UX, Install-Time UX, qmd Integration UX

### Community 43 - "Community 43"
Cohesion: 0.33
Nodes (6): AI-Assisted Development — Compliance & Audit Trail, Audit Trail for AI Actions, Human Oversight Gates, Standards Extraction from Real Code, Project Knowledge Map Template, Standards Directory (AGENTS.md)

### Community 44 - "Community 44"
Cohesion: 0.33
Nodes (6): Chain-of-Thought Framing, Regression Test, ReAct Trace + Session Handoff Example, Compact Prompt, Local Example Prompt, Bugfix RCA Template

### Community 45 - "Community 45"
Cohesion: 0.33
Nodes (6): Advisory Cost Rule Design, API Cost Monitoring, Context Management, LLM & API Cost Management, Model Selection, Token Budget Awareness

### Community 46 - "Community 46"
Cohesion: 0.4
Nodes (4): AdapterContext, AdapterSelections, CompileContext, ToolAdapter

### Community 47 - "Community 47"
Cohesion: 0.4
Nodes (5): Go is Source of Truth, Setup Engine Contracts, Setup Engine Conformance Normalization, MCP Preset Rules, Setup Engine Parity Rules

### Community 48 - "Community 48"
Cohesion: 0.4
Nodes (5): detectProjectStack, Knowledge Map builder, MCP catalog expansion, Post-install automation, Smart Context Builder

### Community 49 - "Community 49"
Cohesion: 0.6
Nodes (5): Store & Errors Constitution, Operation Tracker, Store & Errors Plan, Store & Errors Spec, Zod Single Source of Truth

### Community 50 - "Community 50"
Cohesion: 0.4
Nodes (5): Go Conventions, Error Wrapping with %w, go fmt Non-Negotiable, Sentinel Errors, Table-Driven Tests

### Community 51 - "Community 51"
Cohesion: 0.5
Nodes (3): CompiledFile, CompiledOutput, ToolOverrides

### Community 52 - "Community 52"
Cohesion: 0.5
Nodes (4): Metadata Migration, AGENTS.md Cleanup, Library Replacement Strategy, Root-Only AGENTS.md Target

### Community 53 - "Community 53"
Cohesion: 0.5
Nodes (4): Quality Gates, Implement Prompt, Implement Skill, TDD Loop Skill

### Community 54 - "Community 54"
Cohesion: 0.67
Nodes (3): Codex AGENTS Override Template, Approval Gates for Destructive Operations, Plan Mode for Multi-File Changes

### Community 55 - "Community 55"
Cohesion: 0.67
Nodes (3): Interfaces Over Types, TypeScript Rules, Strict TypeScript

### Community 58 - "Community 58"
Cohesion: 1.0
Nodes (1): Database

### Community 60 - "Community 60"
Cohesion: 1.0
Nodes (2): Session Management & Compaction, Token Discipline

### Community 61 - "Community 61"
Cohesion: 1.0
Nodes (2): Feature Workflow Rules, Tech Debt Workflow Rules

### Community 62 - "Community 62"
Cohesion: 1.0
Nodes (2): Copilot chatmodes, Gemini commands

### Community 63 - "Community 63"
Cohesion: 1.0
Nodes (2): ADR mandatory for refactors, RPI flow

### Community 64 - "Community 64"
Cohesion: 1.0
Nodes (2): LibraryDir dual-path design, go:embed library data

### Community 65 - "Community 65"
Cohesion: 1.0
Nodes (2): Pre-Flight Task Framing, ReAct Trace Pattern

### Community 66 - "Community 66"
Cohesion: 1.0
Nodes (2): build-plugin Command, Claude Plugin Research

### Community 67 - "Community 67"
Cohesion: 1.0
Nodes (2): TypeScript Conventions, Strict Type Discipline

### Community 68 - "Community 68"
Cohesion: 1.0
Nodes (2): Explanatory Output Style, Explanatory Response Style

### Community 69 - "Community 69"
Cohesion: 1.0
Nodes (2): Test Command, Test Failure Summary Extraction

### Community 70 - "Community 70"
Cohesion: 1.0
Nodes (2): Terse Output Style, Terse Response Style

### Community 71 - "Community 71"
Cohesion: 1.0
Nodes (2): Commit Command, Conventional Commits Format

### Community 72 - "Community 72"
Cohesion: 1.0
Nodes (2): Review Command, Review Dimensions

### Community 73 - "Community 73"
Cohesion: 1.0
Nodes (2): Confidence Markers, Uncertainty Markers Template

### Community 74 - "Community 74"
Cohesion: 1.0
Nodes (2): Constitutional Core Principles, Constitution Template

### Community 75 - "Community 75"
Cohesion: 1.0
Nodes (2): Code Style Rule, Consistency Rationale

### Community 85 - "Community 85"
Cohesion: 1.0
Nodes (1): Version 0.1.0

### Community 86 - "Community 86"
Cohesion: 1.0
Nodes (1): Task 005: Library Claude Code Commands

### Community 87 - "Community 87"
Cohesion: 1.0
Nodes (1): Compile-time Scope Awareness

### Community 88 - "Community 88"
Cohesion: 1.0
Nodes (1): Host CLI configuration layouts

### Community 89 - "Community 89"
Cohesion: 1.0
Nodes (1): Phase 1 logging tail CLI

### Community 90 - "Community 90"
Cohesion: 1.0
Nodes (1): Prompts directory

### Community 91 - "Community 91"
Cohesion: 1.0
Nodes (1): Session Handoff Note

### Community 92 - "Community 92"
Cohesion: 1.0
Nodes (1): Post-Task Housekeeping

### Community 93 - "Community 93"
Cohesion: 1.0
Nodes (1): Pre-Flight Task Framing Example

### Community 94 - "Community 94"
Cohesion: 1.0
Nodes (1): Commit Message Pattern Example

### Community 95 - "Community 95"
Cohesion: 1.0
Nodes (1): Spec Template

### Community 96 - "Community 96"
Cohesion: 1.0
Nodes (1): Plan Mode

### Community 97 - "Community 97"
Cohesion: 1.0
Nodes (1): Commit Command

### Community 98 - "Community 98"
Cohesion: 1.0
Nodes (1): Test Command

### Community 99 - "Community 99"
Cohesion: 1.0
Nodes (1): CR Log

### Community 100 - "Community 100"
Cohesion: 1.0
Nodes (1): Automated Formatting

### Community 101 - "Community 101"
Cohesion: 1.0
Nodes (1): Commit Hygiene

### Community 102 - "Community 102"
Cohesion: 1.0
Nodes (1): CopyLibraryDirectory

### Community 103 - "Community 103"
Cohesion: 1.0
Nodes (1): adapter opencode Package

### Community 104 - "Community 104"
Cohesion: 1.0
Nodes (1): adapter claudecode Package

### Community 105 - "Community 105"
Cohesion: 1.0
Nodes (1): adapter codex Package

### Community 106 - "Community 106"
Cohesion: 1.0
Nodes (1): adapter copilot Package

### Community 107 - "Community 107"
Cohesion: 1.0
Nodes (1): adapter gemini Package

### Community 108 - "Community 108"
Cohesion: 1.0
Nodes (1): adapter pi Package

## Ambiguous Edges - Review These
- `Task 015: Docs Knowledge Map and Follow-ups` → `Spec 005: Setup Flow Fixes`  [AMBIGUOUS]
  specs/012-claude-code-deep-setup/tasks/015-docs-knowledge-map-and-followups.md · relation: conceptually_related_to
- `Setup Engine Hardening Design` → `Compile Scope Artifact Parity Research`  [AMBIGUOUS]
  specs/009-compile-scope-artifact-parity/research.md · relation: conceptually_related_to
- `Cost management rule` → `Context Poisoning`  [AMBIGUOUS]
  docs/AI-Agentic-Setup-Templates/docs/rules/AGENTS.md · relation: conceptually_related_to
- `ReAct Trace Pattern` → `Pre-Flight Task Framing`  [AMBIGUOUS]
  specs/prompts/local-examples/preflight-task-framing.md · relation: conceptually_related_to
- `Commit Message Pattern` → `Task Template`  [AMBIGUOUS]
  specs/prompts/local-examples/commit-message-pattern.md · relation: conceptually_related_to
- `Compozy Adaptation Analysis` → `Template Rules and Inventory`  [AMBIGUOUS]
  specs/021-parity-verification/compozy-analysis.md · relation: conceptually_related_to
- `Testing Guidelines (Copilot Instructions)` → `RPI Workflow (Research, Plan, Implement)`  [AMBIGUOUS]
  library/copilot/instructions/tests.instructions.md · relation: conceptually_related_to
- `Testing Guidelines (Copilot Instructions)` → `Agent Coordination (Scout/Planner/Builder/Reviewer/Documenter)`  [AMBIGUOUS]
  library/copilot/instructions/tests.instructions.md · relation: conceptually_related_to
- `Data Domain Skill` → `Testing Rule`  [AMBIGUOUS]
  library/orchestration/skills/domains/data.md · relation: conceptually_related_to
- `DevOps Domain Skill` → `Agent Security Rule`  [AMBIGUOUS]
  library/orchestration/skills/domains/devops.md · relation: conceptually_related_to
- `Junior Mode Skill` → `Self-Improvement Protocol`  [AMBIGUOUS]
  library/orchestration/skills/modes/junior.md · relation: conceptually_related_to
- `ReAct Trace + Session Handoff Example` → `Bugfix RCA Template`  [AMBIGUOUS]
  library/prompts/local-examples/react-trace-and-handoff.md · relation: conceptually_related_to
- `/iterate Skill` → `Bugfix Workflow`  [AMBIGUOUS]
  docs/AI-Agentic-Setup-Templates/docs/change-requests/AGENTS.md · relation: conceptually_related_to
- `Standard Categories` → `Naming Conventions`  [AMBIGUOUS]
  docs/rules/code-style.md · relation: conceptually_related_to

## Knowledge Gaps
- **467 isolated node(s):** `artifactGroup`, `artifactItem`, `CatalogServer`, `Catalog`, `CheckStatus` (+462 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 36`** (9 nodes): `trimExt()`, `PiAdapter`, `.CanRunHeadless()`, `.CompileMCP()`, `.ConfigDir()`, `.ID()`, `.Name()`, `.RunHeadlessValidation()`, `pi.go`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 38`** (8 nodes): `GlobalConfigDir()`, `GlobalSetupDir()`, `IsGlobalSupportedTool()`, `LogUnsupportedGlobalTool()`, `ProjectSetupDir()`, `ResolveGlobalToolTargetDir()`, `WorkspaceSetupDir()`, `globalpaths.go`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 58`** (2 nodes): `better-sqlite3-shim.d.ts`, `Database`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 60`** (2 nodes): `Session Management & Compaction`, `Token Discipline`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 61`** (2 nodes): `Feature Workflow Rules`, `Tech Debt Workflow Rules`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 62`** (2 nodes): `Copilot chatmodes`, `Gemini commands`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 63`** (2 nodes): `ADR mandatory for refactors`, `RPI flow`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 64`** (2 nodes): `LibraryDir dual-path design`, `go:embed library data`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 65`** (2 nodes): `Pre-Flight Task Framing`, `ReAct Trace Pattern`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 66`** (2 nodes): `build-plugin Command`, `Claude Plugin Research`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 67`** (2 nodes): `TypeScript Conventions`, `Strict Type Discipline`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 68`** (2 nodes): `Explanatory Output Style`, `Explanatory Response Style`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 69`** (2 nodes): `Test Command`, `Test Failure Summary Extraction`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 70`** (2 nodes): `Terse Output Style`, `Terse Response Style`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 71`** (2 nodes): `Commit Command`, `Conventional Commits Format`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 72`** (2 nodes): `Review Command`, `Review Dimensions`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 73`** (2 nodes): `Confidence Markers`, `Uncertainty Markers Template`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 74`** (2 nodes): `Constitutional Core Principles`, `Constitution Template`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 75`** (2 nodes): `Code Style Rule`, `Consistency Rationale`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 85`** (1 nodes): `Version 0.1.0`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 86`** (1 nodes): `Task 005: Library Claude Code Commands`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 87`** (1 nodes): `Compile-time Scope Awareness`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 88`** (1 nodes): `Host CLI configuration layouts`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 89`** (1 nodes): `Phase 1 logging tail CLI`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 90`** (1 nodes): `Prompts directory`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 91`** (1 nodes): `Session Handoff Note`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 92`** (1 nodes): `Post-Task Housekeeping`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 93`** (1 nodes): `Pre-Flight Task Framing Example`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 94`** (1 nodes): `Commit Message Pattern Example`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 95`** (1 nodes): `Spec Template`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 96`** (1 nodes): `Plan Mode`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 97`** (1 nodes): `Commit Command`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 98`** (1 nodes): `Test Command`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 99`** (1 nodes): `CR Log`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 100`** (1 nodes): `Automated Formatting`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 101`** (1 nodes): `Commit Hygiene`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 102`** (1 nodes): `CopyLibraryDirectory`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 103`** (1 nodes): `adapter opencode Package`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 104`** (1 nodes): `adapter claudecode Package`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 105`** (1 nodes): `adapter codex Package`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 106`** (1 nodes): `adapter copilot Package`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 107`** (1 nodes): `adapter gemini Package`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 108`** (1 nodes): `adapter pi Package`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `Task 015: Docs Knowledge Map and Follow-ups` and `Spec 005: Setup Flow Fixes`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `Setup Engine Hardening Design` and `Compile Scope Artifact Parity Research`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `Cost management rule` and `Context Poisoning`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `ReAct Trace Pattern` and `Pre-Flight Task Framing`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `Commit Message Pattern` and `Task Template`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `Compozy Adaptation Analysis` and `Template Rules and Inventory`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `Testing Guidelines (Copilot Instructions)` and `RPI Workflow (Research, Plan, Implement)`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
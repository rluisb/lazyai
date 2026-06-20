# Spec 028: Fake Projects Testing & Evidence Collection Plan

## 1. Objective
Plan the execution of `lazyai` to set up fake projects under various configurations, capturing comprehensive evidence of each run. This evidence will be analyzed to identify functional gaps, regressions, and ensure all scenarios work correctly across different environments, AI tools, and presets.

## 2. Testing Dimensions & Matrix

To cover all possible scenarios, the testing matrix will permute across the following dimensions:

### 2.1 Target AI Tools
- `opencode`
- `claude` (Claude Code)
- `cursor`
- `copilot` (GitHub Copilot)
- `gemini` (Gemini CLI)

### 2.2 Initialization Presets
- `minimal`: Bare minimum features, quality gates only.
- `standard`: Recommended baseline (rpiWorkflow, chainOfThought, bugResolution).
- `full`: Maximum guidance and process structure.
- `custom`: Combining specific `--features` and `--disable-features` (e.g., standard but no `treeOfThoughts`).

### 2.3 Execution Scopes
- `project`: Local to the current project directory.
- `global`: Installed to global locations (e.g., `~/.opencode/`, `~/.omp/`).
- `workspace`: Multi-project workspace configurations.

### 2.4 Project Initial States
- **Empty**: Completely empty directory.
- **Git Repo**: Directory with `git init` but no files.
- **Populated Repo**: A realistic codebase (e.g., a basic Node.js or Go project) without AI setup.
- **Legacy Configs**: A project with older `ai-setup` or `LazyAI` configs (to test `migrate` and `import`).

## 3. Operations to Execute

For each matrix combination, the following operations should be simulated:

1. **Environment Validation**: `lazyai-cli doctor --json`
2. **Setup/Init**: `lazyai-cli init --tool <tool> --preset <preset> --scope <scope>`
3. **Configuration**: `lazyai-cli config set <key> <value>`, `lazyai-cli server add ai-memory`
4. **Artifact Creation**: `lazyai-cli create agent test-agent`, `lazyai-cli create skill test-skill`
5. **Session Lifecycle**: `lazyai-cli session start "test task"`, `lazyai-cli session end <id>`
6. **Integrity & Ledger**: `lazyai-cli ledger init`, `lazyai-cli validate skills`, `lazyai-cli validate agents`
7. **Sidecar & Workspace**: `lazyai-cli sidecar init --scope project`, `lazyai-cli workspace add .`

## 4. Evidence Collection Strategy

For every test scenario, a dedicated evidence bundle will be generated to allow deep forensic analysis. 

### 4.1 Artifacts to Collect
- **Execution Logs**: `stdout`, `stderr`, and `exit_code` for every `lazyai-cli` invocation.
- **File System Tree**: `tree -a` output detailing all generated files, directories, and symlinks.
- **Configuration Dumps**: Copies of `.opencode/config.yaml`, generated agent/skill Markdown files, `AGENTS.md`, `RULES.md`, etc.
- **Ledger Logs**: The raw `.specify/ledger.jsonl` to ensure proper action tracking.
- **Health Reports**: `doctor --json` state before and after execution.
- **Project Snapshots**: A `.tar.gz` archive of the entire fake project directory after all operations complete.

### 4.2 Output Structure
The execution framework will output evidences into a structured `evidences/` folder:

```text
evidences/
├── run_metadata.json                 # Overview of all executed matrix runs
├── scenario_001_opencode_minimal_empty/
│   ├── setup_args.json               # CLI args used
│   ├── logs/
│   │   ├── 01_init.log               # stdout/stderr for init
│   │   ├── 02_create_agent.log
│   │   └── 03_doctor.json
│   ├── tree_snapshot.txt             # `tree -a` post-execution
│   ├── file_diffs.patch              # Diff against initial state
│   ├── configs/                      # Copies of important generated config files
│   └── full_project_state.tar.gz     # Complete archive of the fake project
├── scenario_002_claude_standard_git/
│   └── ...
└── ...
```

## 5. Implementation Plan

### Phase 1: Test Runner Scaffold
Develop an automated bash script (`scripts/run_fake_projects.sh`) that:
- Defines arrays for tools, presets, scopes, and project states.
- Creates isolated temporary directories (`/tmp/lazyai-fake-projects/`).
- Seeds the initial state (Empty, Git, Populated, Legacy).

### Phase 2: Execution & Capture Logic
Implement the execution loop to:
- Dynamically build and run `lazyai-cli` commands.
- Wrap executions to capture `$?` (exit code), `stdout`, and `stderr`.
- Produce the output directory structure per scenario.

### Phase 3: Archive & Reporting
- Tarball the state for each scenario.
- Generate an HTML or Markdown summary report (`evidences/report.md`) mapping each scenario to its exit code, indicating which tests passed or failed.
- Include an "Anomalies" section in the report based on parsed `stderr` outputs.

## 6. Edge Cases & Fault Injection Scenarios
1. **Idempotency**: Running `lazyai-cli init` twice in the same project. Evidence must confirm no destructive overwrites or duplicated configurations.
2. **Missing Dependencies**: Running the tests with restricted `$PATH` to simulate environments missing `sqlite3`, `jq`, or `git`. Evidence must capture graceful degradation and clear error messages.
3. **Invalid Frontmatter/Schemas**: Manually injecting malformed YAML into an agent or skill, then executing `validate agents/skills`. Evidence must show validation failures.
4. **Cross-Scope Conflicts**: Running a global initialization followed by a conflicting project initialization. Evidence must capture scoping precedence and overrides.
5. **Disabled Features Overrides**: Using `--disable-features all` and adding specific ones via `--features` to ensure strict exclusion logic holds.

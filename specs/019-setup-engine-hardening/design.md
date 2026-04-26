# Design: Setup Engine Hardening

## 1. Resource Model
The engine treats every setup element as a **Resource**:
- **Tool Target**: The destination environment (e.g., `claude-code`, `opencode`).
- **MCP Server**: A tool provider defined in `catalog.json` (e.g., `codegraph`, `qmd`, `memoria`, `ripgrep`, `filesystem`, `memory`, `orchestrator`).
- **Skill**: A set of instructions and workflows.
- **Agent**: A named configuration comprising a system prompt, associated Skills, and a whitelist of MCP Servers.

## 2. Tool Target Registry
The registry maps targets to their specific config paths and adapter logic:
- **Targets**: `claude-code`, `codex`, `gemini-cli`, `pi`, `opencode`, `copilot-cli`.
- **Adapter Interface**: `Install()`, `Uninstall()`, `GetConfigPath()`, `Validate()`, `Adopt()`.

## 3. Setup Engine Workflow
The setup process follows a linear pipeline to prevent destructive changes:
**Scan $\rightarrow$ Inventory $\rightarrow$ Manifest $\rightarrow$ Apply**

The user-facing selection flow is:
**AI CLI Targets $\rightarrow$ Skills $\rightarrow$ Agents $\rightarrow$ MCP/Tool Servers**

1. **Scan**: Search `~/.ai-setup/`, `.ai/`, and tool-target specific paths (e.g., `~/.claude-code/`).
2. **Inventory**: Compare scanned files against `catalog.json` and project specs.
3. **Manifest**: Generate a proposed state.
4. **Apply**: Implement the changes (copy, link, or modify).

## 4. MCP Server Selection & Presets
MCP servers are selectable resources. To simplify setup, the engine supports **MCP Presets**:
- **Minimal**: Essential core tools only.
- **Recommended**: Balanced set of productivity and discovery tools.
- **Full**: All available tools in the catalog.

Users can further refine selections by manually adding/removing servers from the preset.

## 5. Orchestrator MCP Setup
The `orchestrator` MCP requires a specialized setup routine:
- **Deployment**: Install or build the orchestrator MCP package as needed.
- **Wiring**: Generate and inject MCP config entries into the selected AI CLI tool configs.
- **Verification**: Validate the binary/command path exists and is executable.
- **Smoke Test**: Optional verification that the server starts and responds to a basic request.
- **Integration**: Preserve or adopt existing orchestrator MCP entries if already present.
- **Tracking**: Record orchestrator version and ownership in the setup manifest.

## 6. Pre-existing Setup Absorption
To avoid overwriting user data, the engine implements a **Safe Absorption Policy**:
- **Scan**: Inventory existing config files.
- **Backup**: Create timestamped backups of any file before modification (`.bak`).
- **Conflict States**:
  - `MATCH`: User config matches `ai-setup` spec $\rightarrow$ **Adopt**.
  - `ADOPTABLE`: Entry exists and is compatible; can be moved to managed state.
  - `SENSITIVE`: User config contains secrets/keys not in spec $\rightarrow$ **Merge** (keep keys).
  - `CONFLICT`: User config differs significantly or has conflicting ownership $\rightarrow$ **Prompt User**.
  - `USER_OWNED`: Explicitly marked as non-managed by the user.
  - `MISSING`: Desired resource is not present.
- **Action Options**:
  - `--adopt`: Mark as managed by `ai-setup`.
  - `--import`: Copy into `ai-setup` managed storage.
  - `--skip`: Leave untouched.
  - `--replace`: Overwrite with managed spec.

## 7. Directory & Layout Proposal
### Global Layout (`~/.ai-setup/`)
- `/config/`: Global preferences and target registries.
- `/bin/`: Managed binaries/proxies.
- `/catalog/`: Cached version of `library/mcp/catalog.json`.
- `/backups/`: Storage for absorbed config backups.

### Project Layout (`.ai/`)
- `/agents/`: Directory-based agent definitions (one folder per agent).
  - `AGENT.md`: Prompt and personality (Compozy-style).
  - `mcp.json`: Optional agent-local MCP servers configuration using standard `mcpServers` shape.
  - `/references/`: Optional directory for agent-specific reference documents.
- `/skills/`: Local skill overrides.
- `/env/`: Project-specific environment variables.

### Library Layout (`library/`)
- `/mcp/catalog.json`: Master tool definition.
- `/templates/`: Boilerplate for agent/skill creation.

## 8. Setup Commands Proposal
- `setup --list`: List all supported Tool Targets.
- `setup --scan`: Perform inventory of existing setups.
- `setup --dry-run`: Show proposed changes without applying.
- `setup --tool <name>`: Setup a specific target.
- `setup --all`: Setup all supported targets.
- `setup --global`: Apply global configuration.
- `setup --copy <src> <dst>`: Manually sync a resource.
- `setup --adopt`: Trigger the absorption workflow for existing setups.

## 9. Non-Goals / Out of Scope
- **No Runtime Execution**: No `exec`, `run`, or `prompt` commands.
- **No ACP Session Management**: No state tracking of active AI conversations.
- **No Daemon/Background Processes**: No long-running setup services or default orchestrator daemon control. The AI CLI tool should launch the MCP server via config when needed.
- **No Remote Config Sync**: No cloud-syncing of configs unless via git.

# Research: Engine Hardening for ai-setup

## 1. Compozy Setup Engine Patterns
Compozy treats "setup" as a managed resource lifecycle rather than a one-time script. Key patterns to borrow:
- **Target-Based Registry**: Configuration is mapped to specific "targets" (e.g., Claude Code, Gemini CLI) rather than a generic global state.
- **Resource Model**: Skills, Agents, and MCP servers are treated as resources with defined dependencies and installation requirements.
- **Workspace vs. Global**: Clear separation between `~/.ai-setup/` (global installation/config) and `.ai/` (project-local overrides).
- **Safe Absorption**: A mechanism to scan existing user configurations and "adopt" them into the managed engine rather than overwriting them.
- **Agent-Local Tools**: The ability for an agent definition to specify its own restricted set of MCP servers or tools, ensuring a minimal viable surface area.

## 2. Current ai-setup State
- **Language Mix**: Go (core engine/adapters) and TypeScript (orchestrator/logic).
- **Existing Adapters**: OpenCode, Claude Code, Gemini, Copilot, Codex.
- **Missing**: Pi Go adapter.
- **Asset Structure**:
  - `library/`: Contains core logic and templates.
  - `library/mcp/catalog.json`: Central registry of available MCP servers and CLI tools.
  - `library/orchestration`: Logic for composing agents/chains.
  - `specs/`: Existing planning artifacts.
- **Core Logic**: Primarily focused on the *how* of setup, but lacks a robust *lifecycle* management (scan -> inventory -> adopt -> update).

## 3. Key Gaps
- **Lack of Inventory**: No system to track what is currently installed vs. what is requested in a spec.
- **Brittle Absorption**: Current setup likely overwrites or creates duplicates if the user already has some tools installed manually.
- **Weak Local Scope**: No formalized mechanism for "agent-local" MCP configurations within a project directory.
- **Missing Pi Integration**: No adapter for the Pi target.

## 4. Constraints & Non-Goals
- **NON-GOAL: Runtime Execution**: `ai-setup` must NOT provide an `exec` command, a daemon for session management, or any ACP session runtime. It is a *bootstrapper*, not a *runner*.
- **No State Persistence for Sessions**: It manages the *configuration* of the runtime, not the *state* of the conversation.
- **Tool-Target Specificity**: The engine must only interact with supported Tool Targets (Claude Code, Codex, Gemini CLI, Pi, OpenCode, GitHub Copilot CLI).

## 5. Risk Assessment
- **Conflict Risk**: High. Adopting existing user setups can lead to configuration drift or corrupted config files if not handled via backups.
- **Tool Sprawl**: Managing too many MCP servers can lead to "token bloat" in the prompt. Agent-local restriction is critical.
- **Adapter Divergence**: Different tool targets have wildly different config formats (JSON, YAML, Env vars).

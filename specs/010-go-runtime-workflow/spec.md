# Spec 010: Go Runtime — Workflow Engine

## Chunk 4: YAML Execution, Phase Dispatch, Human Gates

**Status:** Draft  
**Author:** Ricardo Conceicao  
**Date:** 2026-05-26  
**Depends on:** Spec 007 (Foundation), Spec 008 (Session)

---

## §G Goal

Implement the workflow engine that replaces `workflow-run.sh`:
- `load` — Parse YAML workflow definitions
- `run` — Execute workflow phases with actual agent dispatch
- `gate` — Human and automatic gates between phases
- `fallback` — Handle agent failure, test failure, timeout
- `sync` — Sync workflow definitions from YAML to DB
- `list` — List available workflows
- `show` — Show workflow details

**Critical requirement:** Must implement **actual agent dispatch**, not simulation.

---

## §C Constraints

1. **Depends on Spec 007 & 008** — Requires DB and session management
2. **YAML parsing** — Use `gopkg.in/yaml.v3` (already in go.mod)
3. **Variable interpolation** — Support `{GOAL}`, `{COMPLEXITY}`, ternary expressions
4. **Human gates** — Pause execution, prompt user, wait for approval
5. **Actual dispatch** — Must call AI tool (opencode/claude/copilot) to execute agents

---

## §I Interfaces

### §I.1 Workflow Manager

```go
package workflow

import "github.com/rluisb/lazyai/packages/cli/internal/runtime"

// Manager handles workflow operations
type Manager struct {
    db      *runtime.DB
    session *session.Manager
    dispatch dispatch.Dispatcher
}

// NewManager creates a workflow manager
func NewManager(db *runtime.DB, sessionMgr *session.Manager, disp dispatch.Dispatcher) *Manager

// Load parses a YAML workflow definition and stores it
func (m *Manager) Load(yamlPath string) (*Workflow, error)

// Run executes a workflow by name
func (m *Manager) Run(sessionID string, workflowName string, opts RunOptions) (*RunResult, error)

// Sync updates workflow definitions from YAML files
func (m *Manager) Sync(workflowsDir string) error

// List returns available workflows
func (m *Manager) List() ([]Workflow, error)

// Show returns workflow details
func (m *Manager) Show(workflowName string) (*Workflow, error)
```

### §I.2 Data Types

```go
// Workflow represents a workflow definition
type Workflow struct {
    ID          string
    Name        string
    Description string
    Trigger     string
    Config      WorkflowConfig
    Version     int
    CreatedAt   time.Time
    UpdatedAt   *time.Time
    Team        string
    Steps       []WorkflowStep
}

// WorkflowConfig contains mode configurations
type WorkflowConfig struct {
    Modes       map[string]ModeConfig
    DefaultMode string
}

// ModeConfig defines a workflow mode
type ModeConfig struct {
    Phases    []string
    SkipGates bool
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
    Name        string
    Agent       string
    Skill       string
    Mode        string
    Feedforward string
    Gate        *GateConfig
    Metrics     []string
}

// GateConfig defines a human or automatic gate
type GateConfig struct {
    Type        string // "human" or "auto"
    Description string
    Prompt      string
}

// RunResult contains the outcome of a workflow run
type RunResult struct {
    InstanceID   int
    WorkflowName string
    SessionID    string
    Status       string
    CurrentStep  int
    TotalSteps   int
    Result       string
    ErrorMessage string
    StartedAt    time.Time
    CompletedAt  *time.Time
}

// RunOptions configures workflow execution
type RunOptions struct {
    Mode     string
    SkipGates bool
    DryRun   bool
    Context  map[string]string // Variable interpolation context
}
```

---

## §V Invariants

1. **Phase ordering** — Steps execute in definition order, no skipping
2. **Gate enforcement** — Human gates pause execution; auto gates evaluate conditions
3. **Variable interpolation** — All `{VAR}` placeholders resolved before dispatch
4. **Ternary expressions** — `${COND ? 'then' : 'else'}` evaluated at runtime
5. **Fallback cascade** — Agent failure → retry → fallback agent → workflow failure
6. **Ledger integration** — Every phase start/end recorded in ledger

---

## §T Tasks

### Task 1: YAML Parser
**File:** `packages/cli/internal/runtime/workflow/parser.go`

Implement:
- `ParseYAML()` — Parse workflow YAML into Workflow struct
- Support modes (simple/complex) with phase subsets
- Support gates (human/auto) with prompts
- Support variable placeholders in feedforward
- Validate required fields (name, phases, agents)

### Task 2: Variable Interpolation
**File:** `packages/cli/internal/runtime/workflow/interpolate.go`

Implement:
- `Interpolate()` — Replace `{VAR}` with context values
- `EvaluateTernary()` — Parse `${COND ? 'then' : 'else'}`
- Context variables: `GOAL`, `COMPLEXITY`, `CLARIFY_OUTPUT`, etc.
- Phase outputs become inputs to subsequent phases

### Task 3: Workflow Execution
**File:** `packages/cli/internal/runtime/workflow/execute.go`

Implement:
- `Run()` — Execute workflow phases in order
- For each phase:
  1. Interpolate variables
  2. Create workflow instance record
  3. Create workflow step record
  4. **DISPATCH AGENT** (actual execution, not simulation)
  5. Wait for completion
  6. Evaluate gate
  7. Record result to ledger
- Handle human gates (prompt user)
- Handle auto gates (evaluate conditions)
- Handle timeouts

### Task 4: Agent Dispatch Interface
**File:** `packages/cli/internal/runtime/dispatch/dispatcher.go`

```go
package dispatch

// Dispatcher executes agents via AI tools
type Dispatcher interface {
    Dispatch(sessionID string, agent string, task string, mode string) (*DispatchResult, error)
}

// OpenCodeDispatcher implements Dispatcher for OpenCode
type OpenCodeDispatcher struct{}

func (d *OpenCodeDispatcher) Dispatch(sessionID string, agent string, task string, mode string) (*DispatchResult, error) {
    // Execute: opencode run --agent <agent> --mode <mode> <task>
}
```

### Task 5: Fallback Handling
**File:** `packages/cli/internal/runtime/workflow/fallback.go`

Implement:
- `OnAgentFailure()` — Retry with exponential backoff
- `OnTestFailure()` — Update spec, retry, or abort
- `OnTimeout()` — Notify, abort, or extend
- Fallback agent selection from workflow config

### Task 6: Sync
**File:** `packages/cli/internal/runtime/workflow/sync.go`

Implement:
- `Sync()` — Read YAML files, update `workflows` table
- Detect changes (hash comparison)
- Insert new workflows
- Update existing workflows (version bump)
- Delete removed workflows (soft delete)

### Task 7: Tests
**Files:** `packages/cli/internal/runtime/workflow/*_test.go`

Required tests:
- `TestParseYAML` — Valid and invalid YAML
- `TestInterpolate` — Variable replacement, ternary
- `TestExecuteWorkflow` — Full execution with mock dispatcher
- `TestHumanGate` — Pause, prompt, resume
- `TestAutoGate` — Condition evaluation
- `TestFallback` — Agent failure, retry, fallback
- `TestSync` — Insert, update, delete workflows
- `TestConcurrentRuns` — Multiple workflow instances

---

## §B Backward Compatibility

### Bash Command Mapping

| Bash Command | Go Method | Notes |
|-------------|-----------|-------|
| `workflow-run.sh load` | `workflow.Load()` | Parse YAML |
| `workflow-run.sh run` | `workflow.Run()` | **Actual dispatch** |
| `workflow-run.sh sync` | `workflow.Sync()` | YAML → DB sync |
| `workflow-run.sh list` | `workflow.List()` | Same output |
| `workflow-run.sh show` | `workflow.Show()` | Same detail |

### Critical Difference

**Bash:** Simulated execution (TODO at line 461)
```bash
# TODO: Actual dispatch via task tool would go here
# For now, simulate completion
echo "   ✅ Phase $name completed (simulated)"
```

**Go:** Actual dispatch via `dispatch.Dispatcher`
```go
result, err := dispatcher.Dispatch(sessionID, agent, task, mode)
if err != nil {
    return fallback.OnAgentFailure(...)
}
```

---

## §A Acceptance Criteria

1. `go test ./internal/runtime/workflow/...` passes with >90% coverage
2. YAML workflows parse correctly (all 8 existing workflows)
3. Variable interpolation works for all placeholders
4. Human gates pause execution and prompt user
5. Auto gates evaluate conditions correctly
6. Fallback handling retries and selects fallback agents
7. **Actual agent dispatch** executes (not simulates)
8. Ledger records every phase start/end

---

## §N Next Steps

After this spec is approved:
1. Create `packages/cli/internal/runtime/workflow/` directory
2. Create `packages/cli/internal/runtime/dispatch/` directory
3. Implement parser.go (Task 1)
4. Implement interpolate.go (Task 2)
5. Implement execute.go (Task 3)
6. Implement dispatcher.go (Task 4)
7. Implement fallback.go (Task 5)
8. Implement sync.go (Task 6)
9. Write tests (Task 7)
10. PR review and merge

Then proceed to **Chunk 5: Ledger**.

---

## §R References

- Bash workflow engine: `packages/cli/library/fortnite/scripts/workflow-run.sh`
- Bash workflow exec: `packages/cli/library/fortnite/scripts/workflow-exec.sh`
- Current Go workflow: `packages/cli/cmd/workflow.go` (simulated)
- Workflow YAMLs: `packages/cli/library/fortnite/workflows/*.yaml`
- Spec 007 (Foundation): `specs/007-go-runtime-foundation/spec.md`
- Spec 008 (Session): `specs/008-go-runtime-session/spec.md`

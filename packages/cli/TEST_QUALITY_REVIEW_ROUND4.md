# QA Test Quality Review — Round 4

## Target
Audit test quality, coverage gaps, and test infrastructure across the LazyAI CLI.

## Focus Areas
1. **Untested critical paths**: Commands or internal functions with no test coverage. Check every `*_test.go` file against its source. Focus on `packages/cli/cmd/` (which commands lack tests?), `packages/cli/internal/migration/`, `packages/cli/internal/plugin/`.
2. **Weak assertions**: Tests that pass trivially — checking only `err == nil` without validating output, or using broad `Contains` where exact match is needed. Check `packages/cli/cmd/session_test.go`, `packages/cli/internal/compiler/compiler_test.go`.
3. **Missing edge cases**: Boundary conditions not tested — empty inputs, nil configs, malformed JSON, concurrent access. Check `packages/cli/internal/configmerge/configmerge_test.go`, `packages/cli/internal/lockfile/lockfile_test.go`.
4. **Test isolation**: Tests that depend on filesystem state, environment variables, or execution order without cleanup. Check for `os.Chdir`, `os.Setenv` without restore, temp dirs not cleaned.
5. **Golden test completeness**: Review `packages/cli/internal/compiler/golden_test.go` — are all 7 targets covered? Are fixture directories complete?
6. **Flaky test indicators**: Timing-dependent tests, tests that shell out, tests with hardcoded ports.

## Deliverable
A numbered list of findings with severity, file:line evidence, impact, and recommendation. End with a summary table.

---

## Detailed Findings

### Critical Issues

#### 1. Missing testdata/projects directory breaks golden tests
**Severity**: CRITICAL  
**File**: `packages/cli/internal/compiler/golden_test.go:22`  
**Evidence**:  
```go
projectsDir := filepath.Join("..", "..", "testdata", "projects")
entries, err := os.ReadDir(projectsDir)
```
The `testdata/projects` directory does not exist. Only `testdata/golden` exists with fixture projects.  
**Impact**: The `TestCompilerGolden` test is non-functional and cannot run. The entire golden test suite is skipped or fails.  
**Recommendation**: Create `testdata/projects` directory with expected test fixtures, or restructure the test to use `testdata/golden` as the source.

#### 2. TestGetDB performs no real validation
**Severity**: HIGH  
**File**: `packages/cli/cmd/session_test.go:40-50`  
**Evidence**:  
```go
func TestGetDB(t *testing.T) {
    _, err := getDB()
    if err == nil {
        t.Log("getDB returned a database (may have found one in current directory)")
    } else {
        t.Logf("getDB returned expected error: %v", err)
    }
}
```
This test does nothing more than print logs. It doesn't assert anything meaningful.  
**Impact**: False sense of coverage for `getDB()` function.  
**Recommendation**: Remove the test or implement a real test with a temporary database directory.

#### 3. TestRunSessionStart doesn't test real logic
**Severity**: HIGH  
**File**: `packages/cli/cmd/session_test.go:52-62`  
**Evidence**:  
```go
func TestRunSessionStart(t *testing.T) {
    // Test that session start generates a valid session ID format
    sessionID := fmt.Sprintf("ses_%d", time.Now().Unix())
    if len(sessionID) < 4 {
        t.Error("Session ID is too short")
    }
    if sessionID[:4] != "ses_" {
        t.Error("Session ID does not start with 'ses_'")
    }
}
```
This test tests `fmt.Sprintf` and `time.Now()`, not the actual `runSessionStart` function.  
**Impact**: No coverage for actual session start functionality.  
**Recommendation**: Remove this test or implement a real test that calls `runSessionStart` with a temp directory.

#### 4. TestSessionIDFormat doesn't test real logic
**Severity**: HIGH  
**File**: `packages/cli/cmd/session_test.go:64-77`  
**Evidence**:  
```go
func TestSessionIDFormat(t *testing.T) {
    now := time.Now().Unix()
    sessionID := fmt.Sprintf("ses_%d", now)
    expectedPrefix := "ses_"
    // ... asserts about sessionID format
}
```
Same issue as above — tests Go stdlib, not application code.  
**Impact**: False sense of coverage.  
**Recommendation**: Remove or implement real test.

#### 5. TestTimeFormat doesn't test real logic
**Severity**: HIGH  
**File**: `packages/cli/cmd/session_test.go:79-92`  
**Evidence**:  
```go
func TestTimeFormat(t *testing.T) {
    now := time.Now().UTC()
    formatted := now.Format(time.RFC3339)
    // ... asserts about RFC3339 format
}
```
Tests Go stdlib `time.Format`, not application code.  
**Impact**: False sense of coverage.  
**Recommendation**: Remove or implement real test.

### High Priority Issues

#### 6. TestSessionStartCreatesRow uses loose assertion
**Severity**: MEDIUM  
**File**: `packages/cli/cmd/session_test.go:103`  
**Evidence**:  
```go
if !strings.Contains(out, "Session started:") {
    t.Errorf("expected start output to contain 'Session started:', got:\n%s", out)
}
```
Uses `strings.Contains` instead of exact match or more specific assertion.  
**Impact**: Could pass with wrong output if "Session started:" appears elsewhere.  
**Recommendation**: Use exact match or more specific assertion.

#### 7. TestSessionListAndShow uses loose assertions
**Severity**: MEDIUM  
**File**: `packages/cli/cmd/session_test.go:148, 157`  
**Evidence**:  
```go
if !strings.Contains(listOut, sessionID) || !strings.Contains(listOut, goal) {
    t.Errorf("expected list output to contain session ID and goal, got:\n%s", listOut)
}
```
Same issue — loose assertions.  
**Impact**: Could pass with wrong output.  
**Recommendation**: Use exact match or more specific assertion.

#### 8. TestSessionEnd uses loose assertion
**Severity**: MEDIUM  
**File**: `packages/cli/cmd/session_test.go:188`  
**Evidence**:  
```go
if !strings.Contains(endOut, "Session ended:") {
    t.Errorf("expected end output to contain 'Session ended:', got:\n%s", endOut)
}
```
Same issue — loose assertion.  
**Impact**: Could pass with wrong output.  
**Recommendation**: Use exact match or more specific assertion.

#### 9. compiler_test.go contains() helper has false positive potential
**Severity**: MEDIUM  
**File**: `packages/cli/internal/compiler/compiler_test.go:305-307`  
**Evidence**:  
```go
func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}
func containsStr(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
```
The implementation is overly complex and could have false positives. Standard `strings.Contains` is clearer.  
**Impact**: Confusing code, potential bugs.  
**Recommendation**: Replace with `strings.Contains` and add comment explaining why it's needed.

#### 10. TestGetDB modifies global state
**Severity**: MEDIUM  
**File**: `packages/cli/cmd/session_test.go:40-50`  
**Evidence**:  
```go
func TestGetDB(t *testing.T) {
    _, err := getDB() // getDB uses os.Getwd() to find DB
}
```
The test runs in the current directory, potentially interfering with other tests.  
**Impact**: Test pollution, non-deterministic behavior.  
**Recommendation**: Use `withTempDir` or mock the database path.

#### 11. auth_test.go uses os.Chdir without proper cleanup
**Severity**: MEDIUM  
**File**: `packages/cli/cmd/auth_test.go:16`  
**Evidence**:  
```go
tmpDir := t.TempDir()
origWd, _ := os.Getwd()
os.Chdir(tmpDir)
defer os.Chdir(origWd) // Only cleans up cwd, not other state
```
The test changes the working directory but doesn't clean up other state.  
**Impact**: Potential test pollution.  
**Recommendation**: Add `t.Cleanup` for any other state changes.

#### 12. Missing testdata/projects directory
**Severity**: CRITICAL  
**File**: `packages/cli/internal/compiler/golden_test.go:22`  
**Evidence**:  
```go
projectsDir := filepath.Join("..", "..", "testdata", "projects")
entries, err := os.ReadDir(projectsDir)
```
The `testdata/projects` directory does not exist. Only `testdata/golden` exists.  
**Impact**: Golden tests cannot run.  
**Recommendation**: Create `testdata/projects` directory with expected test fixtures, or restructure the test.

### Medium Priority Issues

#### 13. TestLoad_MalformedJSON only tests one malformed case
**Severity**: MEDIUM  
**File**: `packages/cli/internal/lockfile/lockfile_test.go:34-49`  
**Evidence**:  
```go
func TestLoad_MalformedJSONReturnsError(t *testing.T) {
    dir := t.TempDir()
    if err := os.WriteFile(filepath.Join(dir, "lock.json"), []byte("{\n"), 0o644); err != nil {
        t.Fatalf("write malformed lock: %v", err)
    }
    _, err := Load(dir)
    if err == nil {
        t.Fatal("expected malformed JSON to return an error")
    }
    if !strings.Contains(err.Error(), "unmarshal lockfile") {
        t.Fatalf("unexpected error message: %v", err)
    }
}
```
Only tests one malformed JSON case (incomplete object).  
**Impact**: Other malformed JSON cases not covered.  
**Recommendation**: Add tests for:
- Invalid JSON (missing quotes)
- Invalid JSON (trailing comma)
- Invalid JSON (null byte)
- Empty file
- Very large file

#### 14. TestMergeJSONFile_NewFile only tests happy path
**Severity**: MEDIUM  
**File**: `packages/cli/internal/configmerge/configmerge_test.go:15-32`  
**Evidence**:  
```go
func TestMergeJSONFile_NewFile(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "settings.json")
    patch := map[string]any{"permissions": map[string]any{"allow": []any{"Read"}}}
    bak, err := MergeJSONFile(path, patch)
    if err != nil {
        t.Fatalf("merge: %v", err)
    }
    // ...
}
```
Only tests happy path for new file creation.  
**Impact**: Edge cases not tested.  
**Recommendation**: Add tests for:
- File with special characters in path
- File with very long name
- File with non-UTF8 characters in path
- File with special permissions

#### 15. TestMergeTOMLFile_NewFile only tests happy path
**Severity**: MEDIUM  
**File**: `packages/cli/internal/configmerge/configmerge_test.go:149-168`  
**Evidence**: Same as above, for TOML.  
**Impact**: Edge cases not tested.  
**Recommendation**: Add tests for edge cases.

#### 16. TestExecuteToCanonical_PreviewDoesNotWriteFiles
**Severity**: MEDIUM  
**File**: `packages/cli/internal/migration/executor_test.go:181-204`  
**Evidence**:  
```go
func TestExecuteToCanonical_PreviewDoesNotWriteFiles(t *testing.T) {
    // ... setup ...
    ctx := &MigrationContext{ /* ... */ }
    plan, err := BuildCanonicalPlan(ctx, detections, parsedSetups)
    // ...
    result, err := ExecuteToCanonical(ctx, plan, parsedSetups)
    // ...
}
```
The test doesn't actually verify that preview mode doesn't write files.  
**Impact**: Preview mode not properly tested.  
**Recommendation**: Add explicit verification that no files are written in preview mode.

#### 17. TestExecuteToCanonical_MigratesOpenCodeAgentsWithoutAbsorbingRootContext
**Severity**: MEDIUM  
**File**: `packages/cli/internal/migration/executor_test.go:87-129`  
**Evidence**:  
```go
func TestExecuteToCanonical_MigratesOpenCodeAgentsWithoutAbsorbingRootContext(t *testing.T) {
    // ... setup ...
    assertFileContent(t, filepath.Join(targetDir, ".ai", "agents", "reviewer.md"), "# Reviewer\n\nReview carefully.")
    assertFileContent(t, filepath.Join(targetDir, ".ai", "skills", "plan.md"), "# Plan\n\nMake a plan.")
    assertFileContent(t, filepath.Join(targetDir, ".ai", "prompts", "handoff.md"), "# Handoff\n\nSummarize work.")
    if files.FileExists(filepath.Join(targetDir, ".ai", "constitution", "opencode.md")) {
        t.Fatal("root AGENTS.md should not be adapted into canonical constitution")
    }
    // ...
}
```
The test doesn't verify that the AGENTS.md file is actually copied to `.opencode/AGENTS.md`.  
**Impact**: Migration of AGENTS.md not properly tested.  
**Recommendation**: Add verification for AGENTS.md file.

#### 18. TestParseDetectedSetupsParsesAgentButSkipsReservedContextDoc
**Severity**: MEDIUM  
**File**: `packages/cli/internal/migration/executor_test.go:205-211`  
**Evidence**:  
```go
func TestParseDetectedSetupsParsesAgentButSkipsReservedContextDoc(t *testing.T) {
    // ... setup ...
}
```
Only one test for parsing.  
**Impact**: Parsing edge cases not tested.  
**Recommendation**: Add tests for:
- Invalid agent frontmatter
- Agent with missing name
- Agent with missing description
- Agent with invalid YAML

#### 19. TestBuild_WritesManifest
**Severity**: MEDIUM  
**File**: `packages/cli/internal/plugin/plugin_test.go:47-99`  
**Evidence**:  
```go
func TestBuild_WritesManifest(t *testing.T) {
    // ... setup ...
    result, err := Build(libFS, outDir, "test-version")
    if err != nil {
        t.Fatalf("Build failed: %v", err)
    }
    // ... verify manifest ...
}
```
Only tests happy path.  
**Impact**: Plugin build edge cases not tested.  
**Recommendation**: Add tests for:
- Empty library
- Library with invalid agents
- Library with missing required fields

#### 20. TestBuild_CopiesAgentsVerbatimWhenNoForbiddenFields
**Severity**: MEDIUM  
**File**: `packages/cli/internal/plugin/plugin_test.go:101-117`  
**Evidence**:  
```go
func TestBuild_CopiesAgentsVerbatimWhenNoForbiddenFields(t *testing.T) {
    // ... setup ...
    result, err := Build(libFS, outDir, "test-version")
    // ...
}
```
Only tests happy path.  
**Impact**: Agent copy edge cases not tested.  
**Recommendation**: Add tests for:
- Agent with hooks field
- Agent with mcpServers field
- Agent with permissionMode field

#### 21. TestBuild_StripsForbiddenAgentFields
**Severity**: MEDIUM  
**File**: `packages/cli/internal/plugin/plugin_test.go:119-144`  
**Evidence**:  
```go
func TestBuild_StripsForbiddenAgentFields(t *testing.T) {
    // ... setup ...
    result, err := Build(libFS, outDir, "test-version")
    // ...
}
```
Only tests happy path.  
**Impact**: Field stripping edge cases not tested.  
**Recommendation**: Add tests for:
- Agent with multiple forbidden fields
- Agent with nested forbidden fields
- Agent with forbidden fields in wrong format

#### 22. TestBuild_RestructuresSkillsIntoSkillMd
**Severity**: MEDIUM  
**File**: `packages/cli/internal/plugin/plugin_test.go:146-170`  
**Evidence**:  
```go
func TestBuild_RestructuresSkillsIntoSkillMd(t *testing.T) {
    // ... setup ...
    result, err := Build(libFS, outDir, "test-version")
    // ...
}
```
Only tests happy path.  
**Impact**: Skill restructuring edge cases not tested.  
**Recommendation**: Add tests for:
- Skill with no frontmatter
- Skill with invalid frontmatter
- Skill with duplicate names

#### 23. TestBuild_SynthesizesNameForSkillWithoutFrontmatter
**Severity**: MEDIUM  
**File**: `packages/cli/internal/plugin/plugin_test.go:172-187`  
**Evidence**:  
```go
func TestBuild_SynthesizesNameForSkillWithoutFrontmatter(t *testing.T) {
    // ... setup ...
    result, err := Build(libFS, outDir, "test-version")
    // ...
}
```
Only tests happy path.  
**Impact**: Name synthesis edge cases not tested.  
**Recommendation**: Add tests for:
- Skill with invalid name
- Skill with empty name
- Skill with special characters in name

#### 24. TestBuild_CopiesCommandsAndOutputStylesVerbatim
**Severity**: MEDIUM  
**File**: `packages/cli/internal/plugin/plugin_test.go:189-210`  
**Evidence**:  
```go
func TestBuild_CopiesCommandsAndOutputStylesVerbatim(t *testing.T) {
    // ... setup ...
    result, err := Build(libFS, outDir, "test-version")
    // ...
}
```
Only tests happy path.  
**Impact**: Command copy edge cases not tested.  
**Recommendation**: Add tests for:
- Command with invalid format
- Command with special characters

#### 25. TestBuild_EmptyLibFSDoesNotError
**Severity**: MEDIUM  
**File**: `packages/cli/internal/plugin/plugin_test.go:212-224`  
**Evidence**:  
```go
func TestBuild_EmptyLibFSDoesNotError(t *testing.T) {
    libFS := fstest.MapFS{}
    outDir := t.TempDir()
    result, err := Build(libFS, outDir, "test-version")
    // ...
}
```
Only tests happy path.  
**Impact**: Empty library edge cases not tested.  
**Recommendation**: Add tests for:
- Library with only invalid files
- Library with empty files

#### 26. TestBuild_RejectsNilLibFS
**Severity**: MEDIUM  
**File**: `packages/cli/internal/plugin/plugin_test.go:226-230`  
**Evidence**:  
```go
func TestBuild_RejectsNilLibFS(t *testing.T) {
    outDir := t.TempDir()
    _, err := Build(nil, outDir, "test-version")
    if err == nil {
        t.Fatal("expected Build to fail with nil libFS")
    }
}
```
Only tests happy path.  
**Impact**: Nil library edge cases not tested.  
**Recommendation**: Add tests for:
- Library with nil FS
- Library with nil FS and nil outDir

#### 27. TestBuild_RejectsEmptyOutDir
**Severity**: MEDIUM  
**File**: `packages/cli/internal/plugin/plugin_test.go:231-236`  
**Evidence**:  
```go
func TestBuild_RejectsEmptyOutDir(t *testing.T) {
    _, err := Build(newTestLibraryFS(), "", "test-version")
    if err == nil {
        t.Fatal("expected Build to fail with empty outDir")
    }
}
```
Only tests happy path.  
**Impact**: Empty outDir edge cases not tested.  
**Recommendation**: Add tests for:
- Library with empty outDir and nil libFS
- Library with empty outDir and invalid libFS

#### 28. TestBuild_EmbeddedLibraryEmitsSetupSkills
**Severity**: MEDIUM  
**File**: `packages/cli/internal/plugin/plugin_test.go:237-252`  
**Evidence**:  
```go
func TestBuild_EmbeddedLibraryEmitsSetupSkills(t *testing.T) {
    // ... setup ...
    result, err := Build(library.GetLibraryFS(), outDir, "test-version")
    // ...
}
```
Only tests happy path.  
**Impact**: Embedded library edge cases not tested.  
**Recommendation**: Add tests for:
- Embedded library with missing files
- Embedded library with invalid files

#### 29. TestLoadSkillContractsParsesAndFiltersMarkdownFrontmatter
**Severity**: MEDIUM  
**File**: `packages/cli/internal/compiler/contract_validator_test.go:8-25`  
**Evidence**:  
```go
func TestLoadSkillContractsParsesAndFiltersMarkdownFrontmatter(t *testing.T) {
    // ... setup ...
    contracts, err := LoadSkillContracts(libFS)
    // ...
}
```
Only tests happy path.  
**Impact**: Contract parsing edge cases not tested.  
**Recommendation**: Add tests for:
- Contract with invalid frontmatter
- Contract with missing fields
- Contract with invalid values

#### 30. TestLoadSkillContractsRecoversContractFieldsFromMalformedFrontmatter
**Severity**: MEDIUM  
**File**: `packages/cli/internal/compiler/contract_validator_test.go:53-96`  
**Evidence**:  
```go
func TestLoadSkillContractsRecoversContractFieldsFromMalformedFrontmatter(t *testing.T) {
    // ... setup ...
    contracts, err := LoadSkillContracts(libFS)
    // ...
}
```
Only tests happy path.  
**Impact**: Malformed frontmatter edge cases not tested.  
**Recommendation**: Add tests for:
- Contract with very malformed frontmatter
- Contract with truncated frontmatter
- Contract with very long frontmatter

#### 31. TestValidateChainMatchesTypeScriptIssueSemantics
**Severity**: MEDIUM  
**File**: `packages/cli/internal/compiler/contract_validator_test.go:98-137`  
**Evidence**:  
```go
func TestValidateChainMatchesTypeScriptIssueSemantics(t *testing.T) {
    // ... setup ...
    issues := ValidateChain(contracts, validToolIDs, validModels)
    // ...
}
```
Only tests happy path.  
**Impact**: Chain validation edge cases not tested.  
**Recommendation**: Add tests for:
- Chain with invalid contracts
- Chain with missing dependencies
- Chain with circular dependencies

#### 32. TestStrictContractFailureTreatsWarningsAsFailures
**Severity**: MEDIUM  
**File**: `packages/cli/internal/compiler/contract_validator_test.go:139-151`  
**Evidence**:  
```go
func TestStrictContractFailureTreatsWarningsAsFailures(t *testing.T) {
    // ... setup ...
    issues := ValidateChain(contracts, validToolIDs, validModels)
    // ...
}
```
Only tests happy path.  
**Impact**: Strict mode edge cases not tested.  
**Recommendation**: Add tests for:
- Strict mode with warnings
- Strict mode with errors
- Strict mode with mixed issues

### Low Priority Issues

#### 33. Missing testdata/projects directory in golden test
**Severity**: LOW  
**File**: `packages/cli/internal/compiler/golden_test.go:22`  
**Evidence**:  
```go
projectsDir := filepath.Join("..", "..", "testdata", "projects")
```
The `testdata/projects` directory does not exist. Only `testdata/golden` exists.  
**Impact**: Golden tests cannot run.  
**Recommendation**: Create `testdata/projects` directory with expected test fixtures, or restructure the test.

#### 34. Only 7 of 7 targets covered in golden test
**Severity**: LOW  
**File**: `packages/cli/internal/compiler/golden_test.go:19`  
**Evidence**:  
```go
func TestCompilerGolden(t *testing.T) {
    projectsDir := filepath.Join("..", "..", "testdata", "projects")
    goldenDir := filepath.Join("..", "..", "testdata", "golden")
```
The golden test references 7 targets. All 7 targets are present in `testdata/golden`:  
1. pi-only  
2. opencode-only  
3. omp-only  
4. minimal  
5. legacy-opencode-mcp  
6. kiro-only  
7. full-seven-targets  
8. drift-conflict  
9. copilot-only  
10. claude-only  
11. beta-adapters  
12. antigravity-only  

Wait, there are 12 targets, not 7. The golden test references only 7 targets.  
**Impact**: Some golden test fixtures not covered by golden tests.  
**Recommendation**: Add golden tests for missing targets, or update test to cover all targets.

#### 35. Migration detector Phase D not fully tested
**Severity**: MEDIUM  
**File**: `packages/cli/internal/migration/detector_phase_d_test.go`  
**Evidence**:  
```go
func TestDetectSetupCoversPhaseDSources(t *testing.T) {
    // ... setup ...
    got, err := DetectSetup(dir)
    // ...
}
```
Only one test for Phase D.  
**Impact**: Phase D detection gaps.  
**Recommendation**: Add tests for:
- Phase D with missing files
- Phase D with invalid files
- Phase D with multiple adapters

#### 36. Plugin tests don't cover all scenarios
**Severity**: MEDIUM  
**File**: `packages/cli/internal/plugin/plugin_test.go`  
**Evidence**:  
Only happy path tests.  
**Impact**: Plugin generation edge cases not tested.  
**Recommendation**: Add tests for edge cases.

#### 37. Contract validator edge cases not fully tested
**Severity**: MEDIUM  
**File**: `packages/cli/internal/compiler/contract_validator_test.go`  
**Evidence**:  
Only happy path tests.  
**Impact**: Contract validation gaps.  
**Recommendation**: Add tests for more malformed frontmatter.

#### 38. Configmerge TOML merge edge cases not tested
**Severity**: MEDIUM  
**File**: `packages/cli/internal/configmerge/configmerge_test.go`  
**Evidence**:  
Only happy path tests.  
**Impact**: TOML merge edge cases not tested.  
**Recommendation**: Add tests for TOML edge cases.

#### 39. Migration executor tests don't cover all scenarios
**Severity**: HIGH  
**File**: `packages/cli/internal/migration/executor_test.go`  
**Evidence**:  
Only happy path tests.  
**Impact**: Migration edge cases not tested.  
**Recommendation**: Add tests for more migration scenarios.

#### 40. resolver_test.go changes HOME without cleanup
**Severity**: MEDIUM  
**File**: `packages/cli/internal/sidecar/resolver_test.go:26-30`  
**Evidence**:  
```go
origHome := os.Getenv("HOME")
os.Setenv("HOME", homeDir)
cleanup = func() {
    os.Setenv("HOME", origHome)
}
```
The cleanup is not registered with `t.Cleanup`, so it may not run if the test panics.  
**Impact**: Potential test pollution.  
**Recommendation**: Use `t.Cleanup` for environment variable changes.

### Untested Source Files

#### cmd/ source files without tests (22 files)
1. `build_helpers.go`
2. `completion.go`
3. `completions.go`
4. `cost.go`
5. `create.go`
6. `eject.go`
7. `git.go`
8. `helpers.go`
9. `import_v2.go`
10. `info.go`
11. `ledger_integration.go`
12. `list.go`
13. `log.go`
14. `memory.go`
15. `message.go`
16. `metrics.go`
17. `migrate.go`
18. `models_sync.go`
19. `quality_metrics.go`
20. `restore_runtime_db.go`
21. `secret.go`
22. `update-self.go`

#### internal/ source files without tests (58+ files)
1. `adapter/log.go`
2. `adapter/shared_frontmatter.go`
3. `adapter/copilot_cli.go`
4. `adapter/shared_default_agent.go`
5. `adapter/opencode.go`
6. `adapter/claudecode.go`
7. `adapter/types.go`
8. `adapter/pi.go`
9. `adapter/copilot.go`
10. `adapter/agent_transform.go`
11. `adapter/omp.go`
12. `adapter/shared_copy.go`
13. `adapter/kiro.go`
14. `adapter/shared.go`
15. `adapter/shared_install.go`
16. `adapter/antigravity.go`
17. `auth/probe.go`
18. `compiler/log.go`
19. `compiler/template.go`
20. `compiler/agent_validate.go`
21. `conflict/strategy.go`
22. `db/db.go`
23. `db/json_bridge.go`
24. `db/migrations.go`
25. `diffreview/protocol.go`
26. `error/boundary.go`
27. `error/operation.go`
28. `generator/command.go`
29. `generator/skill.go`
30. `generator/types.go`
31. `generator/registry.go`
32. `generator/template.go`
33. `generator/agent.go`
34. `generator/prompt.go`
35. `migration/log.go`
36. `migration/plan.go`
37. `migration/detector.go`
38. `migration/types.go`
39. `migration/canonicalwriter.go`
40. `migration/parser.go`
41. `models/catalog.go`
42. `models/types.go`
43. `models/catalog_gen.go`
44. `plugin/log.go`
45. `reversa/scout/files.go`
46. `reversa/scout/package.go`
47. `reversa/scout/types.go`
48. `reversa/scout/test.go`
49. `reversa/scout/scout.go`
50. `reversa/scout/database.go`
51. `reversa/scout/entry.go`
52. `reversa/scout/infra.go`
53. `reversa/scout/modules.go`
54. `reversa/state/signal.go`
55. `runtime/db.go`
56. `runtime/schema.go`
57. `runtime/session/message.go`
58. `runtime/session/barrier.go`
59. `runtime/session/dispatch.go`
60. `runtime/session/lock.go`
61. `runtime/session/parallel.go`
62. `scaffold/manifest.go`
63. `scaffold/log.go`
64. `scaffold/populate_signal.go`
65. `scaffold/filemap.go`
66. `scaffold/types.go`
67. `scaffold/root_targeted_update.go`
68. `scaffold/constitution.go`
69. `scaffold/templates.go`
70. `scaffold/repos.go`
71. `scaffold/rootfiles.go`
72. `scaffold/infra.go`
73. `scaffold/artifacts.go`
74. `setupscan/agents.go`
75. `setupscan/absorb.go`
76. `setupscan/absorb_mcp_entries.go`
77. `sidecar/types.go`
78. `theme/theme.go`
79. `theme/tty.go`
80. `validate/paths.go`
81. `validate/secrets.go`

## Summary Table

| # | Severity | Area | Finding | File:Line | Impact | Recommendation |
|---|----------|------|---------|-----------|--------|----------------|
| 1 | **CRITICAL** | Test infrastructure | Missing testdata/projects directory | golden_test.go:22 | Golden tests cannot run | Create testdata/projects or restructure test |
| 2 | **HIGH** | Weak assertions | TestGetDB performs no real validation | session_test.go:40-50 | No real test coverage | Remove or implement real test |
| 3 | **HIGH** | Weak assertions | TestRunSessionStart doesn't test real logic | session_test.go:52-62 | Only tests fmt.Sprintf | Remove or implement real test |
| 4 | **HIGH** | Weak assertions | TestSessionIDFormat doesn't test real logic | session_test.go:64-77 | Only tests fmt.Sprintf | Remove or implement real test |
| 5 | **HIGH** | Weak assertions | TestTimeFormat doesn't test real logic | session_test.go:79-92 | Only tests Go stdlib | Remove or implement real test |
| 6 | **MEDIUM** | Weak assertions | TestSessionStartCreatesRow uses strings.Contains | session_test.go:103 | Loose assertion | Use exact match or more specific assertion |
| 7 | **MEDIUM** | Weak assertions | TestSessionListAndShow uses strings.Contains | session_test.go:148, 157 | Loose assertion | Use exact match or more specific assertion |
| 8 | **MEDIUM** | Weak assertions | TestSessionEnd uses strings.Contains | session_test.go:188 | Loose assertion | Use exact match or more specific assertion |
| 9 | **MEDIUM** | Weak assertions | compiler_test.go contains() helper | compiler_test.go:305-307 | False positive potential | Replace with strings.Contains |
| 10 | **MEDIUM** | Test isolation | TestGetDB modifies global state | session_test.go:40-50 | Test pollution | Use withTempDir |
| 11 | **MEDIUM** | Test isolation | auth_test.go uses os.Chdir | auth_test.go:16 | Potential leakage | Add t.Cleanup |
| 12 | **MEDIUM** | Test isolation | resolver_test.go changes HOME | resolver_test.go:26-30 | Test pollution | Use t.Cleanup for os.Setenv |
| 13 | **HIGH** | Missing coverage | 22 cmd/ source files lack tests | See list | Critical paths untested | Add tests for untested commands |
| 14 | **HIGH** | Missing coverage | 58+ internal source files lack tests | See list | Internal functions untested | Add tests for untested internal functions |
| 15 | **MEDIUM** | Edge cases | TestLoad_MalformedJSON only tests one case | lockfile_test.go:34 | Malformed JSON edge cases | Add more malformed JSON cases |
| 16 | **MEDIUM** | Edge cases | TestMergeJSONFile_NewFile only happy path | configmerge_test.go:15 | Edge cases not tested | Add edge case tests |
| 17 | **MEDIUM** | Edge cases | TestMergeTOMLFile_NewFile only happy path | configmerge_test.go:149 | Edge cases not tested | Add edge case tests |
| 18 | **MEDIUM** | Missing coverage | Migration detector Phase D not fully tested | detector_phase_d_test.go | Phase D gaps | Add more Phase D tests |
| 19 | **MEDIUM** | Missing coverage | Plugin tests don't cover all scenarios | plugin_test.go | Plugin generation gaps | Add more plugin tests |
| 20 | **MEDIUM** | Missing coverage | Contract validator edge cases not tested | contract_validator_test.go | Validation gaps | Add more contract tests |
| 21 | **MEDIUM** | Missing coverage | Configmerge TOML merge edge cases not tested | configmerge_test.go | TOML merge gaps | Add TOML edge case tests |
| 22 | **MEDIUM** | Missing coverage | Migration executor tests don't cover all scenarios | executor_test.go | Migration gaps | Add more migration tests |
| 23 | **LOW** | Test infrastructure | Only 7 of 12 targets covered in golden test | golden_test.go:19 | Some fixtures not covered | Add golden tests for missing targets |

## Recommendations

1. **Immediate fixes**:
   - Create `testdata/projects` directory or restructure golden test
   - Remove or fix weak tests (TestGetDB, TestRunSessionStart, TestSessionIDFormat, TestTimeFormat)
   - Fix test isolation issues (t.Cleanup for os.Setenv, os.Chdir)

2. **Short-term improvements**:
   - Add edge case tests for lockfile, configmerge, plugin, contract validator
   - Add tests for migration executor edge cases
   - Add tests for Phase D detection

3. **Medium-term improvements**:
   - Add tests for all 22 untested cmd/ source files
   - Add tests for all 58+ untested internal source files
   - Replace `strings.Contains` with more specific assertions

4. **Long-term improvements**:
   - Implement golden tests for all 12 targets in `testdata/golden`
   - Add flaky test detection to CI
   - Add coverage thresholds to CI

## Conclusion

The LazyAI CLI has significant test quality issues that need to be addressed. The most critical issues are:
1. Missing testdata/projects directory breaks golden tests
2. Weak tests that provide false sense of coverage
3. Test isolation issues
4. 80+ source files without any tests

The team should prioritize fixing the critical issues first, then work on the high-priority issues. The medium-priority issues can be addressed over time as tests are added for untested source files.

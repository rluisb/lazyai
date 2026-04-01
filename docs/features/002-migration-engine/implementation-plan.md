---
title: Migration Engine Implementation Plan
status: draft
date: 2026-04-01
---

# Migration Engine Implementation Plan

## Wave 1: Core Framework (Days 1-3)

### Tasks

#### T001: Create Migration Directory Structure
- [ ] Create `src/migration/` directory
- [ ] Create subdirectories: `registry/`, `parsers/`, `diff/`
- [ ] Set up index.ts exports
- [ ] Update tsconfig if needed

#### T002: Build Detector Module
- [ ] Implement file scanning logic
- [ ] Detection patterns for each adapter
- [ ] Detection result interface
- [ ] Unit tests

#### T003: Create Base Parser Interface
- [ ] Abstract MigrationParser class
- [ ] DetectionResult interface
- [ ] ParsedSetup interface
- [ ] MergeResult interface
- [ ] Documentation

#### T004: Build Parser Registry
- [ ] Registry loader
- [ ] Discovery algorithm (npm/global/local)
- [ ] Parser validation
- [ ] Caching mechanism

#### T005: Create Migration Plan Generator
- [ ] Plan structure
- [ ] Conflict detection
- [ ] Preview generation
- [ ] User prompt integration

### Deliverables
- Core migration framework functional
- Can detect existing setups
- Can generate migration plans

---

## Wave 2: Merger & CLI (Days 4-6)

### Tasks

#### T006: Implement 3-Way Diff Algorithm
- [ ] Myers diff implementation
- [ ] Line-level comparison
- [ ] Conflict marker generation
- [ ] Unit tests with edge cases

#### T007: Build Merger Module
- [ ] Merge strategies: smart, preserve, replace, append
- [ ] Conflict resolution logic
- [ ] Backup creation
- [ ] Merge result reporting

#### T008: Create Executor Module
- [ ] Execute migration plan
- [ ] File operations with rollback
- [ ] Progress reporting
- [ ] Error handling

#### T009: Add Import Command
- [ ] CLI command implementation
- [ ] --preview flag
- [ ] --merge-strategy option
- [ ] Interactive prompts

#### T010: Add Init --migrate Flag
- [ ] Extend init command
- [ ] Auto-detect existing setups
- [ ] Pre-fill wizard answers
- [ ] Migration flow integration

### Deliverables
- Full migration execution pipeline
- CLI commands working
- Can import existing setups

---

## Wave 3: OpenCode Parser (Days 7-9)

### Tasks

#### T011: Implement OpenCode Detection
- [ ] Detect .opencode/ directory
- [ ] Detect AGENTS.md files
- [ ] Detection confidence scoring
- [ ] Tests

#### T012: Parse OpenCode Structure
- [ ] Parse agent definitions
- [ ] Parse command definitions
- [ ] Parse templates
- [ ] Parse project metadata

#### T013: OpenCode Merge Logic
- [ ] Merge strategies for OpenCode files
- [ ] Custom section preservation
- [ ] Conflict resolution
- [ ] Tests

#### T014: Add Doctor --migration-check
- [ ] Compare existing vs ai-setup clean
- [ ] Show drift/differences
- [ ] Generate fix suggestions
- [ ] CLI integration

### Deliverables
- OpenCode migration fully functional
- Doctor command working
- Reference implementation complete

---

## Wave 4: Additional Parsers (Days 10-12)

### Tasks

#### T015: Claude Code Parser
- [ ] Detection logic
- [ ] Parser implementation
- [ ] Merge strategies
- [ ] Tests

#### T016: Pi Parser
- [ ] Detection logic
- [ ] Parser implementation
- [ ] Merge strategies
- [ ] Tests

#### T017: Gemini CLI Parser
- [ ] Detection logic
- [ ] Parser implementation
- [ ] Merge strategies
- [ ] Tests

#### T018: GitHub Copilot Parser
- [ ] Detection logic
- [ ] Parser implementation
- [ ] Merge strategies
- [ ] Tests

### Deliverables
- All 5 adapters supported
- Parser API validated
- Community extension ready

---

## Wave 5: Testing & Documentation (Days 13-14)

### Tasks

#### T019: Sample Setup Creation
- [ ] Create sample OpenCode setup
- [ ] Create sample Claude Code setup
- [ ] Create sample Pi setup
- [ ] Create sample Gemini setup
- [ ] Create sample Copilot setup

#### T020: Integration Tests
- [ ] Full migration scenarios
- [ ] Conflict resolution tests
- [ ] Edge case handling
- [ ] Performance tests

#### T021: Documentation
- [ ] Migration guide
- [ ] Parser API documentation
- [ ] Community extension guide
- [ ] CLI reference

#### T022: Parser Template
- [ ] npm package template
- [ ] Example custom parser
- [ ] Publishing guide

### Deliverables
- Complete test coverage
- Documentation published
- Community templates ready

---

## Total Timeline: 14 days

## Success Criteria
- All 5 adapters can be detected and migrated
- Line-level 3-way merge works correctly
- Parser discovery works for npm/global/local
- Doctor --migration-check shows accurate drift
- 90%+ test coverage
- Documentation complete

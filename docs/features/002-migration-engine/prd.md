---
title: Migration Engine PRD
status: approved
date: 2026-04-01
---

# Migration Engine - Product Requirements Document

## Overview
A professional, community-extensible migration system that intelligently imports existing AI setups into ai-setup format with line-level 3-way merge capabilities.

## Requirements

### CLI Commands
- `ai-setup import [path]` - Primary migration command
- `ai-setup init --migrate` - Init with migration flag  
- `ai-setup migrate [path]` - Alias for import
- `ai-setup doctor --migration-check` - Show drift from clean install

### Parser Discovery (Priority Order)
1. **Project Local**: `./ai-setup/plugins/*/parser.ts`
2. **Global User**: `~/.ai-setup/parsers/*/parser.js`
3. **NPM Packages**: `@ai-setup/parsers-*/dist/parser.js`
4. **Built-in**: `src/migration/parsers/*.ts` (fallback)

### Merge Strategy
**Line-Level 3-Way Merge**:
- Base: ai-setup template
- Ours: Existing user content
- Theirs: User preferences/overrides
- Result: Intelligent line-by-line merge with conflict markers

## Supported Adapters
- OpenCode
- Claude Code
- Pi
- Gemini CLI
- GitHub Copilot
- Custom community parsers

## Test Requirements
- Sample existing setups for each adapter
- Migration verification tests
- Conflict resolution tests

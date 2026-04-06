# {{PROJECT_NAME}}

{{#include fragments/system-context.xml}}

## AI Assistant Configuration

{{TOOL_DESCRIPTION}}

{{#if features.contextEngineering}}
{{#include fragments/context-discipline.md}}
{{/if}}

{{#if features.rpiWorkflow}}
{{#include fragments/rpi-workflow.md}}
{{/if}}

{{#if features.chainOfThought}}
{{#include fragments/reasoning-protocol.md}}
{{/if}}

{{#if features.treeOfThoughts}}
{{#include fragments/decision-protocol.md}}
{{/if}}

{{#if features.qualityGates}}
{{#include fragments/quality-gates.xml}}
{{/if}}

{{#if features.gitConventions}}
{{#include fragments/git-conventions.xml}}
{{/if}}

{{#if features.agentHarness}}
{{#include fragments/agent-harness.md}}
{{/if}}

{{#if features.bugResolution}}
{{#include fragments/bug-resolution.xml}}
{{/if}}

{{TOOL_NOTES}}

## Project-Specific Instructions

{{PROJECT_INSTRUCTIONS}}

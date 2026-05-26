#!/usr/bin/env bash
# rtk-watch.sh — Continuous context monitor for session state
# Triggers on 3 events: start (initial), transition (mid-work), end (handoff/rotate)
# Usage: ./scripts/rtk-watch.sh

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RTK_SDK_ENTRY="${RTK_SDK_ENTRY:-$SCRIPT_DIR/../node_modules/@opencode-ai/sdk/dist/index.js}"

OPENCODE_URL="${OPENCODE_URL:-http://localhost:4096}"
OPENCODE_DIR="${OPENCODE_DIR:-$(pwd)}"

RTK_MEMORY_DIR="${RTK_MEMORY_DIR:-$OPENCODE_DIR/bee-gone/.specify/memory}"
RTK_STATE_DIR="${RTK_STATE_DIR:-$RTK_MEMORY_DIR/.rtk-state}"

RTK_POLL_SECONDS="${RTK_POLL_SECONDS:-30}"
RTK_FAILURE_BACKOFF_SECONDS="${RTK_FAILURE_BACKOFF_SECONDS:-15}"
RTK_CONTEXT_LIMIT="${RTK_CONTEXT_LIMIT:-200000}"
RTK_MESSAGES_LIMIT="${RTK_MESSAGES_LIMIT:-60}"
RTK_DIGEST_TURNS="${RTK_DIGEST_TURNS:-12}"

RTK_CHECKPOINT_AT="${RTK_CHECKPOINT_AT:-0.35}"
RTK_HANDOFF_AT="${RTK_HANDOFF_AT:-0.60}"
RTK_ROTATE_AT="${RTK_ROTATE_AT:-0.80}"

RTK_AUTO_ROTATE="${RTK_AUTO_ROTATE:-0}"

mkdir -p "$RTK_MEMORY_DIR" "$RTK_STATE_DIR"

OPENCODE_URL="$OPENCODE_URL" \
OPENCODE_DIR="$OPENCODE_DIR" \
RTK_MEMORY_DIR="$RTK_MEMORY_DIR" \
RTK_STATE_DIR="$RTK_STATE_DIR" \
RTK_CONTEXT_LIMIT="$RTK_CONTEXT_LIMIT" \
RTK_MESSAGES_LIMIT="$RTK_MESSAGES_LIMIT" \
RTK_DIGEST_TURNS="$RTK_DIGEST_TURNS" \
RTK_CHECKPOINT_AT="$RTK_CHECKPOINT_AT" \
RTK_HANDOFF_AT="$RTK_HANDOFF_AT" \
RTK_ROTATE_AT="$RTK_ROTATE_AT" \
RTK_AUTO_ROTATE="$RTK_AUTO_ROTATE" \
RTK_POLL_SECONDS="$RTK_POLL_SECONDS" \
RTK_FAILURE_BACKOFF_SECONDS="$RTK_FAILURE_BACKOFF_SECONDS" \
RTK_SDK_ENTRY="$RTK_SDK_ENTRY" \
node --input-type=module <<'EOF'
const env = process.env;
const mustBePositiveInt = (name) => {
  const value = Number(env[name]);
  if (!Number.isFinite(value) || value <= 0 || !Number.isInteger(value)) {
    throw new Error(`${name} must be a positive integer. Received: ${env[name]}`);
  }
  return value;
};
const mustBeRatio = (name) => {
  const value = Number(env[name]);
  if (!Number.isFinite(value) || value <= 0 || value >= 1) {
    throw new Error(`${name} must be > 0 and < 1. Received: ${env[name]}`);
  }
  return value;
};
const checkpointAt = mustBeRatio("RTK_CHECKPOINT_AT");
const handoffAt = mustBeRatio("RTK_HANDOFF_AT");
const rotateAt = mustBeRatio("RTK_ROTATE_AT");
if (!(checkpointAt < handoffAt && handoffAt < rotateAt)) {
  throw new Error(`Thresholds must increase: checkpoint < handoff < rotate. Received ${checkpointAt}, ${handoffAt}, ${rotateAt}`);
}
mustBePositiveInt("RTK_POLL_SECONDS");
mustBePositiveInt("RTK_FAILURE_BACKOFF_SECONDS");
mustBePositiveInt("RTK_CONTEXT_LIMIT");
mustBePositiveInt("RTK_MESSAGES_LIMIT");
mustBePositiveInt("RTK_DIGEST_TURNS");
EOF

run_once() {
  OPENCODE_URL="$OPENCODE_URL" \
  OPENCODE_DIR="$OPENCODE_DIR" \
  RTK_MEMORY_DIR="$RTK_MEMORY_DIR" \
  RTK_STATE_DIR="$RTK_STATE_DIR" \
  RTK_CONTEXT_LIMIT="$RTK_CONTEXT_LIMIT" \
  RTK_MESSAGES_LIMIT="$RTK_MESSAGES_LIMIT" \
  RTK_DIGEST_TURNS="$RTK_DIGEST_TURNS" \
  RTK_CHECKPOINT_AT="$RTK_CHECKPOINT_AT" \
  RTK_HANDOFF_AT="$RTK_HANDOFF_AT" \
  RTK_ROTATE_AT="$RTK_ROTATE_AT" \
  RTK_AUTO_ROTATE="$RTK_AUTO_ROTATE" \
  RTK_SDK_ENTRY="$RTK_SDK_ENTRY" \
  node --input-type=module <<'EOF'
import fs from "node:fs/promises";
import path from "node:path";
import { pathToFileURL } from "node:url";

const env = process.env;
const sdk = await import(pathToFileURL(env.RTK_SDK_ENTRY).href);
const { createOpencodeClient } = sdk;

const checkpointAt = Number(env.RTK_CHECKPOINT_AT);
const handoffAt = Number(env.RTK_HANDOFF_AT);
const rotateAt = Number(env.RTK_ROTATE_AT);
const messagesLimit = Number(env.RTK_MESSAGES_LIMIT || 60);
const digestTurns = Number(env.RTK_DIGEST_TURNS || 12);

const client = createOpencodeClient({
  baseUrl: env.OPENCODE_URL,
  directory: env.OPENCODE_DIR,
  responseStyle: "data",
  throwOnError: true,
});

const sessions = await client.session.list({
  query: { directory: env.OPENCODE_DIR },
});

if (!sessions.length) process.exit(0);

const session = [...sessions].sort((a, b) => b.time.updated - a.time.updated)[0];
const stateFile = path.join(env.RTK_STATE_DIR, `${session.id}.json`);

let prev = { band: 0, lastAssistantID: "" };
try {
  prev = JSON.parse(await fs.readFile(stateFile, "utf8"));
} catch {}

const messages = await client.session.messages({
  path: { id: session.id },
  query: { directory: env.OPENCODE_DIR, limit: messagesLimit },
});

if (!messages.length) process.exit(0);

const assistantTurns = messages.filter((m) => m.info.role === "assistant");
const latestAssistantTurn = assistantTurns.at(-1);
if (!latestAssistantTurn) process.exit(0);

const latest = latestAssistantTurn.info;
const inputTokens = latest.tokens?.input ?? 0;
const ratio = inputTokens / Number(env.RTK_CONTEXT_LIMIT || 200000);

const band =
  ratio >= rotateAt ? 3 :
  ratio >= handoffAt ? 2 :
  ratio >= checkpointAt ? 1 :
  0;

const lastErrorTurn = [...assistantTurns].reverse().find((m) => m.info.error);
const lastError = lastErrorTurn
  ? JSON.stringify(lastErrorTurn.info.error)
  : "none";

const digest = messages
  .slice(-digestTurns)
  .flatMap((m) => {
    const role = m.info.role;
    return m.parts
      .filter((p) => p.type === "text" && p.text?.trim())
      .slice(0, 1)
      .map((p) => `- ${role}: ${p.text.trim().replace(/\s+/g, " ").slice(0, 280)}`);
  })
  .join("\n") || "- none";

const now = new Date().toISOString();
const pct = (ratio * 100).toFixed(1);
const worktree = latest.path?.root || latest.path?.cwd || session.directory;

const checkpointBody = `---
type: checkpoint
session_id: ${session.id}
updated_at: ${now}
worktree_path: ${worktree}
context_ratio: ${ratio.toFixed(4)}
source: rtk-watch.sh
---

## Goal
- ${session.title || "(untitled session)"}

## Current state
- session_id: ${session.id}
- directory: ${session.directory}
- estimated_context_usage: ${pct}%
- latest_assistant_message: ${latest.id}

## Decisions with why
- not derived automatically; run /rtk-check for semantic checkpointing

## Errors / blockers
- ${lastError}

## Next action
- if usage >= ${(handoffAt * 100).toFixed(0)}% run /rtk-handoff
- if usage >= ${(rotateAt * 100).toFixed(0)}% start a fresh child session

## Resume prompt
- Resume from checkpoint__${session.id}.md. Reconfirm the goal, latest blocker, and next concrete action.
`;

const checkpointFile = path.join(env.RTK_MEMORY_DIR, `checkpoint__${session.id}.md`);

if (band >= 1 && (prev.band < 1 || prev.lastAssistantID !== latest.id)) {
  await fs.writeFile(checkpointFile, checkpointBody, "utf8");
  console.log(`[rtk] checkpoint updated: ${path.basename(checkpointFile)} (${pct}%)`);
}

let handoffFile = null;

if (band >= 2 && (prev.band < 2 || prev.lastAssistantID !== latest.id)) {
  handoffFile = path.join(
    env.RTK_MEMORY_DIR,
    `handoff__${session.id}__${now.replace(/[:.]/g, "-")}.md`,
  );

  const handoffBody = `---
type: handoff
from_session: ${session.id}
created_at: ${now}
worktree_path: ${worktree}
context_ratio: ${ratio.toFixed(4)}
reason: threshold
---

## Task
- ${session.title || "(untitled session)"}

## Current state
- session_id: ${session.id}
- directory: ${session.directory}
- estimated_context_usage: ${pct}%
- latest_assistant_message: ${latest.id}

## Decisions and why
- not derived automatically; confirm with /rtk-check or /rtk-handoff

## Errors / blockers
- ${lastError}

## First step for next session
- Run /rtk-resume with this handoff, then run /task-context before new work.

## Resume prompt
- Use this handoff as the current source of truth. Recover objective, blocker, and next action. Do not re-explore settled context unless evidence is missing.
`;

  await fs.writeFile(handoffFile, handoffBody, "utf8");
  console.log(`[rtk] handoff written: ${path.basename(handoffFile)} (${pct}%)`);
}

if (band >= 3 && env.RTK_AUTO_ROTATE === "1" && (prev.band < 3 || prev.lastAssistantID !== latest.id)) {
  const child = await client.session.create({
    body: {
      parentID: session.id,
      title: `Resume: ${session.title || session.id}`,
    },
    query: { directory: env.OPENCODE_DIR },
  });

  const handoffText = handoffFile
    ? await fs.readFile(handoffFile, "utf8")
    : checkpointBody;

  await client.session.prompt({
    path: { id: child.id },
    query: { directory: env.OPENCODE_DIR },
    body: {
      noReply: true,
      parts: [
        {
          type: "text",
          text: handoffText,
        },
      ],
    },
  });

  console.log(`[rtk] child session created: ${child.id}`);
}

await fs.writeFile(
  stateFile,
  JSON.stringify(
    {
      band,
      lastAssistantID: latest.id,
      ratio,
      updatedAt: now,
    },
    null,
    2,
  ),
  "utf8",
);
EOF
}

while true; do
  if ! run_once; then
    printf '[rtk] watch iteration failed; retrying in %ss\n' "$RTK_FAILURE_BACKOFF_SECONDS" >&2
    sleep "$RTK_FAILURE_BACKOFF_SECONDS"
    continue
  fi

  sleep "$RTK_POLL_SECONDS"
done

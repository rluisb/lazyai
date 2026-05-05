# Policies for Claude Code — RPI Gate Enforcement
# Uses Claude Code-specific tool names and signal context
package rpi.claude

import future.keywords.if
import future.keywords.in

# ─── RPI Phase Gate Enforcement ───

# Default: deny writes to src/ unless plan is attested by a human
default allow_src_write := false

allow_src_write if {
	signal.plan_attested
	signal.author_is_human
}

# Require human review for any src/ write without approved plan
require_review if {
	input.tool == "write_file"
	regex.match(input.path, "src/")
	not allow_src_write
}

# Require human review for commits >20 lines without gate attestation
require_review if {
	input.tool == "git_commit"
	signal.lines_changed > 20
	not signal.gate_attested
}

# Require human review for plan.md creation if spec.md doesn't exist (Article III)
require_review if {
	input.tool == "write_file"
	regex.match(input.path, "specs/.*/plan\\.md$")
	not signal.spec_exists
}

# Require human review for src/ refactors without ADR (Article III)
require_review if {
	input.tool == "write_file"
	regex.match(input.path, "src/")
	signal.lines_changed > 100
	not signal.adr_exists
}

# ─── Hard Blocks ───

# Block: push to main/master
block if {
	input.tool == "git_push"
	signal.branch == "main"
}

# Block: push to main/master (alternative branch name)
block if {
	input.tool == "git_push"
	signal.branch == "master"
}

# Block: force push anywhere
block if {
	input.tool == "git_push"
	some arg in input.args
	arg == "--force"
}

# Block: commit with --no-verify unless approved
block if {
	input.tool == "git_commit"
	some arg in input.args
	arg == "--no-verify"
	not signal.gate_attested
}

# ─── Test-First Enforcement (Article II) ───

# Warn: writing to src/ without test file existing
warn if {
	input.tool == "write_file"
	regex.match(input.path, "src/.*\\.(ts|js|py|rs|go)$")
	not signal.has_tests
	not regex.match(input.path, ".*\\.test\\.")
	not regex.match(input.path, ".*\\.spec\\.")
}

# ─── Anti-Overengineering (Article VI) ───

# Warn: file exceeds 300 lines
warn if {
	input.tool == "write_file"
	signal.file_too_large
}

# ─── Feedback Messages ───

feedback := "⛔ GATE REQUIRED: Write blocked. A plan.md with 'Human Gate: APPROVED' written by a human must exist before writing to src/. Create or approve the plan first." if {
	not allow_src_write
	input.tool == "write_file"
	regex.match(input.path, "src/")
}

feedback := "⛔ GATE REQUIRED: Commit blocked. Changes >20 lines require gate attestation. Ensure plan.md exists with 'Human Gate: APPROVED'." if {
	input.tool == "git_commit"
	signal.lines_changed > 20
	not signal.gate_attested
}

feedback := "⛔ BLOCKED: Cannot push directly to main. Use a feature branch and create a pull request." if {
	input.tool == "git_push"
	signal.branch in {"main", "master"}
}

feedback := "⚠️ TDD: Test file not found for this source file. Write the test first (Article II)." if {
	input.tool == "write_file"
	regex.match(input.path, "src/.*\\.(ts|js|py|rs|go)$")
	not signal.has_tests
}

feedback := "⚠️ Article VI: File exceeds 300 lines. Consider splitting." if {
	input.tool == "write_file"
	signal.file_too_large
}

feedback := "⛔ DOCS FIRST: plan.md requires spec.md to exist first (Article III)." if {
	input.tool == "write_file"
	regex.match(input.path, "specs/.*/plan\\.md$")
	not signal.spec_exists
}

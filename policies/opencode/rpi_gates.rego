# Policies for OpenCode — RPI Gate Enforcement
# Uses OpenCode-specific tool names. OpenCode's mode permissions (write: false
# in plan mode) provide an additional enforcement layer.
package rpi.opencode

import future.keywords.if
import future.keywords.in

# ─── RPI Phase Gate Enforcement ───

default allow_src_write := false

allow_src_write if {
	signal.plan_attested
	signal.author_is_human
}

# Block: write to src/ without approved plan
block if {
	input.tool == "Write"
	regex.match(input.path, "src/")
	not allow_src_write
}

# Block: commit >20 lines without gate attestation
block if {
	input.tool == "Bash"
	regex.match(input.command, "git commit")
	signal.lines_changed > 20
	not signal.gate_attested
}

# Block: push to main/master
block if {
	input.tool == "Bash"
	regex.match(input.command, "git push.*(main|master)")
}

# Block: force push
block if {
	input.tool == "Bash"
	regex.match(input.command, "git push.*--force")
}

# Block: plan.md creation if spec.md doesn't exist (Article III)
block if {
	input.tool == "Write"
	regex.match(input.path, "specs/.*/plan\\.md$")
	not signal.spec_exists
}

# ─── Test-First Enforcement (Article II) ───

warn if {
	input.tool == "Write"
	regex.match(input.path, "src/.*\\.(ts|js|py|rs|go)$")
	not signal.has_tests
	not regex.match(input.path, ".*\\.test\\.")
	not regex.match(input.path, ".*\\.spec\\.")
}

# ─── Anti-Overengineering (Article VI) ───

warn if {
	input.tool == "Write"
	signal.file_too_large
}

# ─── Feedback Messages ───

feedback := "⛔ GATE REQUIRED: Write blocked. plan.md with 'Human Gate: APPROVED' must exist and be written by a human." if {
	input.tool == "Write"
	regex.match(input.path, "src/")
	not allow_src_write
}

feedback := "⛔ BLOCKED: Cannot push directly to main. Use a feature branch and create a pull request." if {
	input.tool == "Bash"
	regex.match(input.command, "git push.*(main|master)")
}

feedback := "⚠️ TDD: Write the failing test first before implementing (Article II)." if {
	input.tool == "Write"
	regex.match(input.path, "src/.*\\.(ts|js|py|rs|go)$")
	not signal.has_tests
}

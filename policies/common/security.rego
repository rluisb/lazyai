# Common Security Policies — Applied across all harnesses
package security.common

import future.keywords.if

# ─── Secret Detection ───

# Block: secrets in written content
block if {
	input.tool == "write_file"
	re_match(input.content, "(?i)(api[_-]?key|apikey|secret[_-]?key|token|password|passwd|credential|private[_-]?key)\\s*[:=]\\s*['\"][^'\"]{8,}['\"]")
}

# Block: AWS access keys
block if {
	input.tool == "write_file"
	re_match(input.content, "AKIA[0-9A-Z]{16}")
}

# Block: GitHub tokens
block if {
	input.tool == "write_file"
	re_match(input.content, "ghp_[0-9a-zA-Z]{36}")
}

# Block: Stripe keys
block if {
	input.tool == "write_file"
	re_match(input.content, "(sk_live_|pk_live_|rk_live_)[0-9a-zA-Z]{24,}")
}

# Block: Google API keys
block if {
	input.tool == "write_file"
	re_match(input.content, "AIza[0-9A-Za-z\\-_]{35}")
}

# Block: JWT tokens written to code
block if {
	input.tool == "write_file"
	re_match(input.content, "eyJ[a-zA-Z0-9_-]{10,}\\.[a-zA-Z0-9_-]{10,}\\.[a-zA-Z0-9_-]{10,}")
}

# Block: Private SSH keys
block if {
	input.tool == "write_file"
	re_match(input.content, "-----BEGIN (RSA|DSA|EC|OPENSSH) PRIVATE KEY-----")
}

feedback := "⛔ SECURITY: Secret detected in content. Remove the hardcoded secret and use environment variables or a secrets manager." if {
	block
}

# ─── SQL Injection Prevention ───

# Block: string-concatenated SQL queries
block if {
	input.tool == "write_file"
	regex.match(input.content, "(?i)(SELECT|INSERT|UPDATE|DELETE|CREATE TABLE|ALTER TABLE|DROP TABLE).*\\+.*")
}

# Block: f-string SQL queries (Python)
block if {
	input.tool == "write_file"
	regex.match(input.content, "(?i)f[\"']\\s*(SELECT|INSERT|UPDATE|DELETE)")
}

# Block: template literal SQL queries (JavaScript/TypeScript)
block if {
	input.tool == "write_file"
	regex.match(input.content, "(?i)`\\s*(SELECT|INSERT|UPDATE|DELETE)")
	not regex.match(input.content, "\\$[0-9]")  # Allow if parameterized
}

feedback := "⛔ SECURITY: SQL injection pattern detected. Use parameterized queries instead of string concatenation. For TypeScript: use $1, $2. For Python: use %s with cursor.execute(query, params)." if {
	block
}

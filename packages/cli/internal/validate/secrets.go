package validate

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/jsonc"
)

// secretPattern is a named, high-confidence inline-secret matcher. The set is
// deliberately narrow to avoid false positives on placeholder/example values.
type secretPattern struct {
	name string
	re   *regexp.Regexp
}

var secretPatterns = []secretPattern{
	{"aws-access-key-id", regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)},
	{"private-key-block", regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH |DSA |PGP )?PRIVATE KEY-----`)},
	{"openai-key", regexp.MustCompile(`\bsk-[A-Za-z0-9]{20,}\b`)},
	{"github-token", regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{30,}\b`)},
	{"slack-token", regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{10,}\b`)},
	{"google-api-key", regexp.MustCompile(`\bAIza[0-9A-Za-z\-_]{35}\b`)},
}

// secretEnvKeyRe matches env var names whose value should be a reference, not
// an inline literal (e.g. API_KEY, AUTH_TOKEN, DB_PASSWORD, CLIENT_SECRET).
var secretEnvKeyRe = regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password|passwd|credential|access[_-]?key|private[_-]?key)`)

// envRefRe detects ${VAR} / $VAR references, which are safe (no inline value).
var envRefRe = regexp.MustCompile(`\$\{?[A-Za-z_][A-Za-z0-9_]*\}?`)

// secretSeverity maps the active profile to the severity of an inline secret:
// error under team, warning under personal (FR-010).
func secretSeverity(p Profile) Severity {
	if p == ProfileTeam {
		return SeverityError
	}
	return SeverityWarning
}

// scanSecrets detects inline secrets across markdown assets and MCP env values
// (FR-010). References of the form ${VAR}/$VAR are always treated as safe.
func scanSecrets(aiDir string, profile Profile, r *Report) {
	sev := secretSeverity(profile)
	scanAssetSecrets(aiDir, sev, r)
	scanMCPEnvSecrets(aiDir, sev, r)
}

func scanAssetSecrets(aiDir string, sev Severity, r *Report) {
	_ = filepath.WalkDir(aiDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		switch ext {
		case ".md", ".mdc", ".txt", ".sh", ".bash", ".ts", ".js", ".yaml", ".yml":
		default:
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		text := string(content)
		rel := relForReport(aiDir, path)
		for _, pat := range secretPatterns {
			if pat.re.MatchString(text) {
				r.add(rel, "secret", "", sev, "possible inline secret (%s)", pat.name)
			}
		}
		return nil
	})
}

func scanMCPEnvSecrets(aiDir string, sev Severity, r *Report) {
	path := filepath.Join(aiDir, "mcp.json")
	if _, err := os.Stat(path); err != nil {
		alt := filepath.Join(aiDir, "mcp.jsonc")
		if _, altErr := os.Stat(alt); altErr != nil {
			return
		}
		path = alt
	}
	doc, err := jsonc.ReadJSONCFile(path)
	if err != nil {
		return // structure errors are reported by validateMCP
	}
	servers := mcpServers(doc)
	if servers == nil {
		return
	}
	rel := relForReport(aiDir, path)
	for name, raw := range servers {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		env, ok := entry["env"].(map[string]any)
		if !ok {
			continue
		}
		for key, val := range env {
			value, ok := val.(string)
			if !ok {
				continue
			}
			value = strings.TrimSpace(value)
			if value == "" || envRefRe.MatchString(value) {
				continue // empty or a ${VAR}/$VAR reference: safe
			}
			if matchesSecretPattern(value) {
				r.add(rel, "secret", "", sev, "server %q env %q has an inline secret value", name, key)
				continue
			}
			if secretEnvKeyRe.MatchString(key) {
				r.add(rel, "secret", "", sev, "server %q env %q holds an inline literal; use a ${VAR} reference", name, key)
			}
		}
	}
}

func matchesSecretPattern(value string) bool {
	for _, pat := range secretPatterns {
		if pat.re.MatchString(value) {
			return true
		}
	}
	return false
}

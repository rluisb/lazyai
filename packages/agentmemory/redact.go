package agentmemory

import "regexp"

var redactionPatterns = []struct {
	re *regexp.Regexp
	to string
}{
	{regexp.MustCompile(`sk-[a-zA-Z0-9]{32,}`), `sk-REDACTED`},
	{regexp.MustCompile(`Bearer [a-zA-Z0-9._\-]+`), `Bearer REDACTED`},
	{regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----[a-zA-Z0-9/\n+=]*-----END [A-Z ]*PRIVATE KEY-----`), `[REDACTED PRIVATE KEY]`},
	{regexp.MustCompile(`(?m)^([A-Za-z0-9_]*(?:_KEY|_SECRET|_TOKEN)=).*`), `${1}REDACTED`},
}

// Redact replaces known secret shapes with deterministic placeholders.
func Redact(input string) string {
	output := input
	for _, pattern := range redactionPatterns {
		output = pattern.re.ReplaceAllString(output, pattern.to)
	}
	return output
}

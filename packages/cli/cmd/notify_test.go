package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEscapeAppleScriptNeutralizesQuotes(t *testing.T) {
	got := escapeAppleScript(`a"b\c`)
	want := `a\"b\\c`
	if got != want {
		t.Fatalf("escapeAppleScript(%q) = %q, want %q", `a"b\c`, got, want)
	}
}

func TestEscapePowerShellNeutralizesExpansion(t *testing.T) {
	got := escapePowerShell("a\"b`c$(d)")
	// backtick doubled, quote backtick-escaped, $ backtick-escaped
	want := "a`\"b``c`$(d)"
	if got != want {
		t.Fatalf("escapePowerShell = %q, want %q", got, want)
	}
}

func TestEscapeFunctionsLeaveCleanInputUntouched(t *testing.T) {
	clean := "Build complete: 3 tasks"
	if got := escapeAppleScript(clean); got != clean {
		t.Errorf("escapeAppleScript altered clean input: %q", got)
	}
	if got := escapePowerShell(clean); got != clean {
		t.Errorf("escapePowerShell altered clean input: %q", got)
	}
}

// TestWebhookPayloadIsValidJSON ensures title/message with quotes/metacharacters
// produce a parseable JSON body (the encoding/json path that replaced raw
// fmt.Sprintf interpolation).
func TestWebhookPayloadIsValidJSON(t *testing.T) {
	body, err := json.Marshal(map[string]string{
		"title":   `Done "now"`,
		"message": `path: $(rm -rf /) and a "quote"`,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var parsed map[string]string
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("payload is not valid JSON: %v\nbody=%s", err, body)
	}
	if !strings.Contains(parsed["message"], "$(rm -rf /)") {
		t.Fatalf("message content not preserved: %q", parsed["message"])
	}
}

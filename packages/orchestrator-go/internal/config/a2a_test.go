package config

import (
	"strings"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/types"
)

func TestParseValidConfig(t *testing.T) {
	cfg, err := Parse([]byte(validConfigJSON()))
	if err != nil {
		t.Fatalf("Parse valid config: %v", err)
	}
	if cfg.Version != 1 || cfg.Execution.Mode != types.ExecutionA2A {
		t.Fatalf("unexpected config version/mode: %+v", cfg)
	}
	if got := cfg.Providers["local"].Auth.Type; got != AuthBearerEnv {
		t.Fatalf("expected bearerEnv auth, got %q", got)
	}
	if !cfg.Agents["builder"].IsEnabled() {
		t.Fatalf("expected builder agent to be enabled")
	}
}

func TestParseRejectsInvalidMode(t *testing.T) {
	_, err := Parse([]byte(strings.Replace(validConfigJSON(), `"mode":"a2a"`, `"mode":"embedded"`, 1)))
	assertConfigError(t, err, "invalid execution mode")
}

func TestParseRejectsMissingProviderReference(t *testing.T) {
	_, err := Parse([]byte(strings.Replace(validConfigJSON(), `"provider":"local"`, `"provider":"missing"`, 1)))
	assertConfigError(t, err, "unknown provider")
}

func TestParseRejectsInvalidProviderURL(t *testing.T) {
	_, err := Parse([]byte(strings.Replace(validConfigJSON(), `"https://a2a.example.test/rpc"`, `"ftp://a2a.example.test/rpc"`, 1)))
	assertConfigError(t, err, "http or https")
}

func TestParseRejectsUnknownTool(t *testing.T) {
	_, err := Parse([]byte(strings.Replace(validConfigJSON(), `"opencode"`, `"vim"`, 1)))
	assertConfigError(t, err, "unsupported tool")
}

func TestParseRejectsInlineSecretLikeFields(t *testing.T) {
	input := strings.Replace(validConfigJSON(), `"env":"A2A_TOKEN"`, `"token":"literal-secret"`, 1)
	_, err := Parse([]byte(input))
	assertConfigError(t, err, "inline secret-like field")
}

func assertConfigError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error %q did not contain %q", err.Error(), want)
	}
}

func validConfigJSON() string {
	return `{
  "version":1,
  "execution":{"mode":"a2a"},
  "providers":{
    "local":{
      "endpoint":"https://a2a.example.test/rpc",
      "auth":{"type":"bearerEnv","env":"A2A_TOKEN"}
    }
  },
  "agents":{
    "builder":{"provider":"local","enabled":true,"tools":["opencode","claude-code"]}
  }
}`
}

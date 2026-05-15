package agentmemory

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRedactPatterns(t *testing.T) {
	input := strings.Join([]string{
		"openai=sk-abcdefghijklmnopqrstuvwxyz1234567890",
		"auth=Bearer abc.def-123",
		"-----BEGIN RSA PRIVATE KEY-----",
		"YWJjZGVmZw==",
		"-----END RSA PRIVATE KEY-----",
		"SERVICE_TOKEN=supersecret",
	}, "\n")

	got := Redact(input)
	for _, secret := range []string{"abcdefghijklmnopqrstuvwxyz1234567890", "abc.def-123", "YWJjZGVmZw==", "supersecret"} {
		if strings.Contains(got, secret) {
			t.Fatalf("Redact() leaked %q in %q", secret, got)
		}
	}
	for _, want := range []string{"sk-REDACTED", "Bearer REDACTED", "[REDACTED PRIVATE KEY]", "SERVICE_TOKEN=REDACTED"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Redact() missing %q in %q", want, got)
		}
	}
}

func TestRedactPreservesJSONStructure(t *testing.T) {
	input := `{"token":"Bearer abc.def-123","key":"sk-abcdefghijklmnopqrstuvwxyz1234567890","clean":"value"}`
	got := Redact(input)

	var decoded map[string]string
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("redacted JSON is invalid: %v; %s", err, got)
	}
	if decoded["token"] != "Bearer REDACTED" || decoded["key"] != "sk-REDACTED" || decoded["clean"] != "value" {
		t.Fatalf("decoded redacted JSON = %+v", decoded)
	}
}

func TestRedactCleanInputNoop(t *testing.T) {
	input := "clean content with no secrets"
	if got := Redact(input); got != input {
		t.Fatalf("Redact(clean) = %q, want %q", got, input)
	}
}

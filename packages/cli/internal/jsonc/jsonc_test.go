package jsonc

import (
	"testing"
)

func TestStripComments_SingleLineComments(t *testing.T) {
	t.Parallel()

	input := []byte(`{
  // This is a comment
  "name": "test",
  "value": 42 // inline comment
}`)

	result := StripComments(input)

	if result == nil {
		t.Fatal("result is nil")
	}

	// Should not contain comment text
	s := string(result)
	if contains(s, "This is a comment") {
		t.Error("single-line comment not stripped")
	}
	if contains(s, "inline comment") {
		t.Error("inline comment not stripped")
	}
	// Should still contain the actual values
	if !contains(s, `"name": "test"`) {
		t.Error("actual content was removed")
	}
	if !contains(s, `"value": 42`) {
		t.Error("actual content was removed")
	}
}

func TestStripComments_MultiLineComments(t *testing.T) {
	t.Parallel()

	input := []byte(`{
  /* multi
     line
     comment */
  "key": "value"
}`)

	result := StripComments(input)
	s := string(result)

	if contains(s, "multi") {
		t.Error("multi-line comment not stripped")
	}
	if !contains(s, `"key": "value"`) {
		t.Error("actual content was removed")
	}
}

func TestStripComments_PreservesStrings(t *testing.T) {
	t.Parallel()

	input := []byte(`{
  "url": "https://example.com/path",
  "text": "use // for comments"
}`)

	result := StripComments(input)
	s := string(result)

	if !contains(s, "https://example.com/path") {
		t.Error("URL was incorrectly stripped as a comment")
	}
	if !contains(s, "use // for comments") {
		t.Error("string with // was incorrectly stripped")
	}
}

func TestParseJSONC(t *testing.T) {
	t.Parallel()

	input := []byte(`{
  // Configuration file
  "name": "ai-setup",
  "version": "1.0.0", /* version info */
  "features": {
    "enabled": true
  }
}`)

	result, err := ParseJSONC(input)
	if err != nil {
		t.Fatalf("ParseJSONC: %v", err)
	}

	if result["name"] != "ai-setup" {
		t.Errorf("name = %v, want ai-setup", result["name"])
	}
	if result["version"] != "1.0.0" {
		t.Errorf("version = %v, want 1.0.0", result["version"])
	}
}

func TestParseJSONC_InvalidJSON(t *testing.T) {
	t.Parallel()

	input := []byte(`{invalid json}`)
	_, err := ParseJSONC(input)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestStripComments_NoComments(t *testing.T) {
	t.Parallel()

	input := []byte(`{"key": "value"}`)
	result := StripComments(input)

	if string(result) != `{"key": "value"}` {
		t.Errorf("stripped = %q, want %q", string(result), `{"key": "value"}`)
	}
}

// helper to avoid importing strings
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package cmd

import (
	"testing"
)

func TestFormatTaskID(t *testing.T) {
	cases := []struct {
		id       int
		expected string
	}{
		{1, "task_1"},
		{42, "task_42"},
		{0, "task_0"},
		{-1, "task_-1"},
	}
	for _, c := range cases {
		got := formatTaskID(c.id)
		if got != c.expected {
			t.Errorf("formatTaskID(%d) = %q, want %q", c.id, got, c.expected)
		}
	}
}

func TestParseTaskID(t *testing.T) {
	cases := []struct {
		input       string
		wantID      int
		wantErr     bool
		wantErrMsg  string
	}{
		{"task_1", 1, false, ""},
		{"task_42", 42, false, ""},
		{"task_0", 0, false, ""},
		{"task_-1", -1, false, ""},
		{"task", 0, true, "invalid task ID format: task"},
		{"task_", 0, true, "invalid task ID format: task_"},
		{"task_abc", 0, true, "invalid task ID format: task_abc"},
		{"task_1_2", 0, true, "invalid task ID format: task_1_2"},
		{"", 0, true, "invalid task ID format: "},
	}
	for _, c := range cases {
		got, err := parseTaskID(c.input)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseTaskID(%q) expected error, got nil", c.input)
				continue
			}
			if err.Error() != c.wantErrMsg {
				t.Errorf("parseTaskID(%q) error = %q, want %q", c.input, err.Error(), c.wantErrMsg)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseTaskID(%q) unexpected error: %v", c.input, err)
			continue
		}
		if got != c.wantID {
			t.Errorf("parseTaskID(%q) = %d, want %d", c.input, got, c.wantID)
		}
	}
}

func TestFormatParseRoundTrip(t *testing.T) {
	for _, id := range []int{1, 2, 99, 12345} {
		formatted := formatTaskID(id)
		parsed, err := parseTaskID(formatted)
		if err != nil {
			t.Fatalf("round-trip for %d: parse error: %v", id, err)
		}
		if parsed != id {
			t.Errorf("round-trip for %d: got %d", id, parsed)
		}
	}
}

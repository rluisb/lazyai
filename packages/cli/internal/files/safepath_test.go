package files

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeJoin(t *testing.T) {
	root := t.TempDir()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errSub  string
	}{
		{
			name:  "normal relative path succeeds",
			input: "agents/reviewer.md",
		},
		{
			name:  "empty string succeeds (joins to root)",
			input: "",
		},
		{
			name:  "path with embedded .. that doesn't escape succeeds",
			input: "a/../b",
		},
		{
			name:  "nested relative path succeeds",
			input: ".ai/agents/reviewer.md",
		},
		{
			name:    "traversal escape rejected",
			input:   "../../../etc/passwd",
			wantErr: true,
			errSub:  "escapes root",
		},
		{
			name:    "absolute path rejected",
			input:   "/etc/passwd",
			wantErr: true,
			errSub:  "absolute paths are not allowed",
		},
		{
			name:    "windows rooted path rejected",
			input:   `\Windows\System32`,
			wantErr: true,
			errSub:  "absolute paths are not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeJoin(root, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("SafeJoin(%q) expected error, got %q", tt.input, got)
				}
				if tt.errSub != "" && !strings.Contains(err.Error(), tt.errSub) {
					t.Fatalf("SafeJoin(%q) error %q does not contain %q", tt.input, err, tt.errSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("SafeJoin(%q) unexpected error: %v", tt.input, err)
			}
			expected := filepath.Join(root, tt.input)
			if got != expected {
				t.Fatalf("SafeJoin(%q) = %q, want %q", tt.input, got, expected)
			}
		})
	}
}

func TestSafeJoin_EscapeVariants(t *testing.T) {
	root := t.TempDir()

	escapeCases := []string{
		"../",
		"..",
		"../secret",
		"a/../../b",
		"foo/../../bar",
	}
	for _, input := range escapeCases {
		t.Run(input, func(t *testing.T) {
			_, err := SafeJoin(root, input)
			if err == nil {
				t.Fatalf("SafeJoin(%q) expected error", input)
			}
		})
	}
}

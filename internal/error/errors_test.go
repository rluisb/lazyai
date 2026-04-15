package error

import (
	"errors"
	"fmt"
	"testing"
)

func TestFileNotFound(t *testing.T) {
	t.Parallel()

	err := FileNotFound("/some/path")
	if err.Code != ErrFileNotFound {
		t.Errorf("Code = %q, want %q", err.Code, ErrFileNotFound)
	}
	if err.Message == "" {
		t.Error("Message is empty")
	}
	if err.Context["path"] != "/some/path" {
		t.Errorf("Context[path] = %v, want /some/path", err.Context["path"])
	}
}

func TestFilePermission(t *testing.T) {
	t.Parallel()

	err := FilePermission("/some/file", "read")
	if err.Code != ErrFilePermission {
		t.Errorf("Code = %q, want %q", err.Code, ErrFilePermission)
	}
}

func TestFileCorrupt(t *testing.T) {
	t.Parallel()

	cause := fmt.Errorf("bad format")
	err := FileCorrupt("/some/corrupt", cause)
	if err.Code != ErrFileCorrupt {
		t.Errorf("Code = %q, want %q", err.Code, ErrFileCorrupt)
	}
	if err.Cause != cause {
		t.Error("Cause not set correctly")
	}
}

func TestUserCancelled(t *testing.T) {
	t.Parallel()

	err := UserCancelled()
	if err.Code != ErrUserCancelled {
		t.Errorf("Code = %q, want %q", err.Code, ErrUserCancelled)
	}
}

func TestInvalidInput(t *testing.T) {
	t.Parallel()

	err := InvalidInput("bad value", map[string]any{"field": "name"})
	if err.Code != ErrInvalidInput {
		t.Errorf("Code = %q, want %q", err.Code, ErrInvalidInput)
	}
	if err.Context["field"] != "name" {
		t.Errorf("Context[field] = %v, want name", err.Context["field"])
	}
}

func TestIsUserError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code ErrorCode
		want bool
	}{
		{ErrUserCancelled, true},
		{ErrInvalidInput, true},
		{ErrConflictUnresolved, true},
		{ErrFileNotFound, false},
		{ErrFilePermission, false},
		{ErrManifestCorrupt, false},
		{ErrUnknown, false},
	}

	for _, tt := range tests {
		err := &AiSetupError{Code: tt.code, Message: "test"}
		got := err.IsUserError()
		if got != tt.want {
			t.Errorf("IsUserError(%q) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestExitCode(t *testing.T) {
	t.Parallel()

	cancelled := &AiSetupError{Code: ErrUserCancelled}
	if cancelled.ExitCode() != 0 {
		t.Errorf("ExitCode(USER_CANCELLED) = %d, want 0", cancelled.ExitCode())
	}

	fileNotFound := &AiSetupError{Code: ErrFileNotFound}
	if fileNotFound.ExitCode() != 1 {
		t.Errorf("ExitCode(FILE_NOT_FOUND) = %d, want 1", fileNotFound.ExitCode())
	}
}

func TestAsAiSetupError(t *testing.T) {
	t.Parallel()

	t.Run("direct AiSetupError", func(t *testing.T) {
		t.Parallel()
		original := FileNotFound("test.txt")
		var target *AiSetupError
		ok := AsAiSetupError(original, &target)
		if !ok {
			t.Fatal("AsAiSetupError returned false")
		}
		if target.Code != ErrFileNotFound {
			t.Errorf("Code = %q, want %q", target.Code, ErrFileNotFound)
		}
	})

	t.Run("wrapped AiSetupError", func(t *testing.T) {
		t.Parallel()
		original := FileNotFound("wrapped.txt")
		wrapped := fmt.Errorf("outer: %w", original)
		var target *AiSetupError
		ok := AsAiSetupError(wrapped, &target)
		if !ok {
			t.Fatal("AsAiSetupError returned false for wrapped error")
		}
		if target.Code != ErrFileNotFound {
			t.Errorf("Code = %q, want %q", target.Code, ErrFileNotFound)
		}
	})

	t.Run("non-AiSetupError", func(t *testing.T) {
		t.Parallel()
		plainErr := errors.New("plain error")
		var target *AiSetupError
		ok := AsAiSetupError(plainErr, &target)
		if ok {
			t.Error("AsAiSetupError should return false for plain error")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()
		var target *AiSetupError
		ok := AsAiSetupError(nil, &target)
		if ok {
			t.Error("AsAiSetupError should return false for nil")
		}
	})
}

func TestError_ImplementsError(t *testing.T) {
	t.Parallel()

	var err error = FileNotFound("test.txt")
	if err.Error() == "" {
		t.Error("Error() returned empty string")
	}
}

func TestUnwrap(t *testing.T) {
	t.Parallel()

	cause := fmt.Errorf("root cause")
	err := FileCorrupt("test.txt", cause)
	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Error("Unwrap() did not return the cause")
	}
}

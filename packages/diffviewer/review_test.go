package diffviewer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateReviewRequest(t *testing.T) {
	t.Parallel()

	title := "Review AGENTS.md update"

	tests := []struct {
		name    string
		req     ReviewRequest
		wantErr string
	}{
		{
			name: "valid request with all fields",
			req: ReviewRequest{
				Version: 1,
				Title:   &title,
				Files: []FileDiff{
					{
						Path:           "AGENTS.md",
						CurrentContent: "old\n",
						NewContent:     "new\n",
					},
				},
			},
		},
		{
			name: "valid request with minimal fields",
			req: ReviewRequest{
				Version: 1,
				Files: []FileDiff{
					{
						Path:           "AGENTS.md",
						CurrentContent: "",
						NewContent:     "new\n",
					},
				},
			},
		},
		{
			name: "unsupported version",
			req: ReviewRequest{
				Version: 2,
				Files: []FileDiff{
					{Path: "AGENTS.md", CurrentContent: "old", NewContent: "new"},
				},
			},
			wantErr: "version",
		},
		{
			name: "missing file path",
			req: ReviewRequest{
				Version: 1,
				Files: []FileDiff{
					{CurrentContent: "old", NewContent: "new"},
				},
			},
			wantErr: "path",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateReviewRequest(tt.req)
			assertValidationError(t, err, tt.wantErr)
		})
	}
}

func TestValidateReviewResponse(t *testing.T) {
	t.Parallel()

	message := "Review confirmed."

	tests := []struct {
		name    string
		resp    ReviewResponse
		wantErr string
	}{
		{
			name: "invalid action",
			resp: ReviewResponse{
				Version: 1,
				Status:  ReviewStatusConfirmed,
				Resolutions: []Resolution{
					{Path: "AGENTS.md", Action: Action("replace")},
				},
			},
			wantErr: "action",
		},
		{
			name: "valid confirmed response",
			resp: ReviewResponse{
				Version: 1,
				Status:  ReviewStatusConfirmed,
				Resolutions: []Resolution{
					{Path: "AGENTS.md", Action: ActionAccept},
					{Path: "README.md", Action: ActionSkip},
				},
				Message: &message,
			},
		},
		{
			name: "valid cancelled response",
			resp: ReviewResponse{
				Version:     1,
				Status:      ReviewStatusCancelled,
				Resolutions: []Resolution{},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateReviewResponse(tt.resp)
			assertValidationError(t, err, tt.wantErr)
		})
	}
}

func TestReviewContractExampleFixturesValidate(t *testing.T) {
	t.Parallel()

	fixtures := []struct {
		name string
		kind string
	}{
		{name: "request-multi-file.json", kind: "request"},
		{name: "request-single-file.json", kind: "request"},
		{name: "response-cancelled.json", kind: "response"},
		{name: "response-confirmed.json", kind: "response"},
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(filepath.Join(contractExampleDir(t), fixture.name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			switch fixture.kind {
			case "request":
				var req ReviewRequest
				if err := json.Unmarshal(data, &req); err != nil {
					t.Fatalf("unmarshal request fixture: %v", err)
				}
				if err := ValidateReviewRequest(req); err != nil {
					t.Fatalf("validate request fixture: %v", err)
				}
			case "response":
				var resp ReviewResponse
				if err := json.Unmarshal(data, &resp); err != nil {
					t.Fatalf("unmarshal response fixture: %v", err)
				}
				if err := ValidateReviewResponse(resp); err != nil {
					t.Fatalf("validate response fixture: %v", err)
				}
			default:
				t.Fatalf("unsupported fixture kind %q", fixture.kind)
			}
		})
	}
}

func assertValidationError(t *testing.T, err error, wantSubstring string) {
	t.Helper()

	if wantSubstring == "" {
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		return
	}

	if err == nil {
		t.Fatalf("expected error containing %q", wantSubstring)
	}
	if !strings.Contains(err.Error(), wantSubstring) {
		t.Fatalf("expected error containing %q, got %v", wantSubstring, err)
	}
}

func contractExampleDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test filename")
	}

	return filepath.Join(filepath.Dir(filename), "..", "..", "specs", "features", "interactive-diff-review", "contracts", "examples")
}

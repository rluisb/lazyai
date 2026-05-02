package diffviewer

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestRunCLIValidRequestProcessesThroughInjectedReviewer(t *testing.T) {
	t.Parallel()

	stdin := strings.NewReader(`{
		"version": 1,
		"files": [
			{"path":"AGENTS.md","currentContent":"old\nline","newContent":"new\nline"}
		]
	}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	wantResponse := ReviewResponse{
		Version: reviewContractVersion,
		Status:  ReviewStatusConfirmed,
		Resolutions: []Resolution{
			{Path: "AGENTS.md", Action: ActionAccept},
		},
	}

	var gotViews []ConflictView
	factory := func(views []ConflictView) Reviewer {
		gotViews = append([]ConflictView(nil), views...)
		return reviewerFunc(func() (ReviewResponse, error) {
			stderr.WriteString("tui rendered on stderr\n")
			return wantResponse, nil
		})
	}

	code := RunCLI(stdin, &stdout, &stderr, factory)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	if len(gotViews) != 1 {
		t.Fatalf("expected reviewer to receive 1 view, got %d", len(gotViews))
	}
	if gotViews[0].FilePath != "AGENTS.md" {
		t.Fatalf("expected view path AGENTS.md, got %q", gotViews[0].FilePath)
	}
	if !reflect.DeepEqual(gotViews[0].CurrentLines, []string{"old", "line"}) {
		t.Fatalf("unexpected current lines: %#v", gotViews[0].CurrentLines)
	}
	if !reflect.DeepEqual(gotViews[0].NewLines, []string{"new", "line"}) {
		t.Fatalf("unexpected new lines: %#v", gotViews[0].NewLines)
	}

	assertStdoutJSONOnly(t, stdout.Bytes(), wantResponse)
	if !strings.Contains(stderr.String(), "tui rendered on stderr") {
		t.Fatalf("expected TUI output on stderr, got %q", stderr.String())
	}
}

func TestRunCLIInvalidRequestWritesErrorToStderrAndJSONToStdout(t *testing.T) {
	t.Parallel()

	stdin := strings.NewReader(`{"version":2,"files":[]}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunCLI(stdin, &stdout, &stderr, func([]ConflictView) Reviewer {
		t.Fatal("reviewer should not run for invalid request")
		return nil
	})
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "version") {
		t.Fatalf("expected validation error on stderr, got %q", stderr.String())
	}

	var response ReviewResponse
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &response); err != nil {
		t.Fatalf("stdout must contain valid JSON response bytes, got %q: %v", stdout.String(), err)
	}
	if response.Status != ReviewStatusCancelled {
		t.Fatalf("expected cancelled error response, got %#v", response)
	}
	if response.Message == nil || !strings.Contains(*response.Message, "version") {
		t.Fatalf("expected error response message to contain validation error, got %#v", response.Message)
	}
}

func assertStdoutJSONOnly(t *testing.T, output []byte, want ReviewResponse) {
	t.Helper()

	trimmed := bytes.TrimSpace(output)
	if !json.Valid(trimmed) {
		t.Fatalf("stdout must contain only valid JSON response bytes, got %q", string(output))
	}
	if bytes.Contains(trimmed, []byte("tui rendered")) {
		t.Fatalf("stdout contains TUI output: %q", string(output))
	}

	var got ReviewResponse
	if err := json.Unmarshal(trimmed, &got); err != nil {
		t.Fatalf("unmarshal stdout response: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected response\n got: %#v\nwant: %#v", got, want)
	}
}

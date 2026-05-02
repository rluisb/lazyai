package wizard

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestShouldDelegateReviewThreshold(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		conflicts []conflict.Conflict
		want      bool
	}{
		{
			name: "single file small diff stays inline",
			conflicts: []conflict.Conflict{
				newReviewConflict("AGENTS.md", "current\nshared", "new\nshared"),
			},
			want: false,
		},
		{
			name: "multiple files delegates",
			conflicts: []conflict.Conflict{
				newReviewConflict("AGENTS.md", "current", "new"),
				newReviewConflict("README.md", "current", "new"),
			},
			want: true,
		},
		{
			name: "single file large diff delegates",
			conflicts: []conflict.Conflict{
				newReviewConflict("AGENTS.md", twentyNumberedLines("current"), twentyNumberedLines("new")),
			},
			want: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := ShouldDelegateReview(tc.conflicts); got != tc.want {
				t.Fatalf("ShouldDelegateReview() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBinaryDiffReviewerMissingBinaryFallsBackInline(t *testing.T) {
	t.Parallel()

	inline := &recordingDiffReviewClient{
		resolutions: []ConflictResolution{{Path: "AGENTS.md", Action: ReviewActionDeny}},
	}
	reviewer := BinaryDiffReviewer{
		Inline:     inline,
		Runner:     failDiffReviewRunner{t: t},
		IsTerminal: func() bool { return true },
		BinaryResolver: func(binaryName string) (string, error) {
			if binaryName != defaultDiffViewerBinary {
				t.Fatalf("binary resolver got %q, want %q", binaryName, defaultDiffViewerBinary)
			}
			return "", errors.New("missing diffviewer")
		},
	}

	got, err := reviewer.RunReview([]conflict.Conflict{
		newReviewConflict("AGENTS.md", twentyNumberedLines("current"), twentyNumberedLines("new")),
	})
	if err != nil {
		t.Fatalf("RunReview: %v", err)
	}
	if !inline.called {
		t.Fatal("inline reviewer was not called")
	}
	if !reflect.DeepEqual(got, inline.resolutions) {
		t.Fatalf("resolutions = %#v, want %#v", got, inline.resolutions)
	}
}

func TestBinaryDiffReviewerBelowThresholdUsesInline(t *testing.T) {
	t.Parallel()

	inline := &recordingDiffReviewClient{
		resolutions: []ConflictResolution{{Path: "small.md", Action: ReviewActionAccept}},
	}
	reviewer := BinaryDiffReviewer{
		Inline: inline,
		Runner: failDiffReviewRunner{t: t},
		BinaryResolver: func(binaryName string) (string, error) {
			t.Fatalf("binary resolver should not run below threshold; got %q", binaryName)
			return "", nil
		},
		IsTerminal: func() bool {
			t.Fatal("terminal detector should not run below threshold")
			return true
		},
	}

	got, err := reviewer.RunReview([]conflict.Conflict{
		newReviewConflict("small.md", "current\nshared", "new\nshared"),
	})
	if err != nil {
		t.Fatalf("RunReview: %v", err)
	}
	if !inline.called {
		t.Fatal("inline reviewer was not called")
	}
	if !reflect.DeepEqual(got, inline.resolutions) {
		t.Fatalf("resolutions = %#v, want %#v", got, inline.resolutions)
	}
}

func TestBinaryDiffReviewerNonTerminalUsesInline(t *testing.T) {
	t.Parallel()

	inline := &recordingDiffReviewClient{
		resolutions: []ConflictResolution{{Path: "large.md", Action: ReviewActionSkip}},
	}
	reviewer := BinaryDiffReviewer{
		Inline:     inline,
		Runner:     failDiffReviewRunner{t: t},
		IsTerminal: func() bool { return false },
		BinaryResolver: func(binaryName string) (string, error) {
			t.Fatalf("binary resolver should not run in non-terminal mode; got %q", binaryName)
			return "", nil
		},
	}

	got, err := reviewer.RunReview([]conflict.Conflict{
		newReviewConflict("large.md", twentyNumberedLines("current"), twentyNumberedLines("new")),
	})
	if err != nil {
		t.Fatalf("RunReview: %v", err)
	}
	if !inline.called {
		t.Fatal("inline reviewer was not called")
	}
	if !reflect.DeepEqual(got, inline.resolutions) {
		t.Fatalf("resolutions = %#v, want %#v", got, inline.resolutions)
	}
}

func TestInlineDiffReviewerReturnsPerFileDecisions(t *testing.T) {
	t.Parallel()

	prompter := &recordingInlinePrompter{
		actions: []ReviewAction{ReviewActionAccept, ReviewActionSkip},
	}
	reviewer := InlineDiffReviewer{Prompter: prompter}

	got, err := reviewer.RunReview([]conflict.Conflict{
		newReviewConflict("one.md", "current", "new"),
		newReviewConflict("two.md", "current", "new"),
	})
	if err != nil {
		t.Fatalf("RunReview: %v", err)
	}

	want := []ConflictResolution{
		{Path: "one.md", Action: ReviewActionAccept},
		{Path: "two.md", Action: ReviewActionSkip},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resolutions = %#v, want %#v", got, want)
	}
	if !reflect.DeepEqual(prompter.paths, []string{"one.md", "two.md"}) {
		t.Fatalf("prompt paths = %#v", prompter.paths)
	}
}

func TestRunPhase3NonInteractiveSkipsConflicts(t *testing.T) {
	t.Parallel()

	result, action, err := RunPhase3([]conflict.Conflict{
		newReviewConflict("large.md", twentyNumberedLines("current"), twentyNumberedLines("new")),
	}, true)
	if err != nil {
		t.Fatalf("RunPhase3: %v", err)
	}
	if action != PhaseContinue {
		t.Fatalf("action = %v, want %v", action, PhaseContinue)
	}
	if result.Strategy != types.ConflictStrategySkip {
		t.Fatalf("strategy = %q, want %q", result.Strategy, types.ConflictStrategySkip)
	}
	if len(result.Resolutions) != 0 {
		t.Fatalf("resolutions = %#v, want none", result.Resolutions)
	}
}

func TestLargeDiffMissingBinaryFallsBackWithoutCrash(t *testing.T) {
	t.Parallel()

	inline := &recordingDiffReviewClient{
		resolutions: []ConflictResolution{{Path: "large.md", Action: ReviewActionSkip}},
	}
	reviewer := BinaryDiffReviewer{
		Inline:         inline,
		Runner:         failDiffReviewRunner{t: t},
		IsTerminal:     func() bool { return true },
		BinaryResolver: func(string) (string, error) { return "", errors.New("missing diffviewer") },
	}

	got, err := reviewer.RunReview([]conflict.Conflict{
		newReviewConflict("large.md", twentyNumberedLines("current"), twentyNumberedLines("new")),
	})
	if err != nil {
		t.Fatalf("RunReview: %v", err)
	}
	if !inline.called {
		t.Fatal("inline reviewer was not called")
	}
	if !reflect.DeepEqual(got, inline.resolutions) {
		t.Fatalf("resolutions = %#v, want %#v", got, inline.resolutions)
	}
}

func TestConflictStrategyForReviewAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action ReviewAction
		want   types.ConflictStrategy
	}{
		{name: "accept maps to backup and replace", action: ReviewActionAccept, want: types.ConflictStrategyBackupAndReplace},
		{name: "deny maps to skip", action: ReviewActionDeny, want: types.ConflictStrategySkip},
		{name: "skip maps to align", action: ReviewActionSkip, want: types.ConflictStrategyAlign},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := ConflictStrategyForReviewAction(tc.action); got != tc.want {
				t.Fatalf("ConflictStrategyForReviewAction(%q) = %q, want %q", tc.action, got, tc.want)
			}
		})
	}
}

func newReviewConflict(path, current, next string) conflict.Conflict {
	return conflict.Conflict{Path: path, CurrentContent: []byte(current), NewContent: []byte(next)}
}

type recordingDiffReviewClient struct {
	called      bool
	resolutions []ConflictResolution
	err         error
}

func (r *recordingDiffReviewClient) RunReview(conflicts []conflict.Conflict) ([]ConflictResolution, error) {
	r.called = true
	if r.err != nil {
		return nil, r.err
	}
	return r.resolutions, nil
}

type failDiffReviewRunner struct {
	t *testing.T
}

func (r failDiffReviewRunner) RunDiffReview(_ context.Context, binaryPath string, _ []byte) ([]byte, error) {
	r.t.Fatalf("delegated diffviewer runner should not run; binaryPath=%q", binaryPath)
	return nil, nil
}

type recordingInlinePrompter struct {
	actions []ReviewAction
	paths   []string
}

func (p *recordingInlinePrompter) PromptReview(c conflict.Conflict, index, total int) (ReviewAction, error) {
	p.paths = append(p.paths, c.Path)
	if index < 1 || index > total {
		return "", errors.New("invalid prompt index")
	}
	if len(p.paths) > len(p.actions) {
		return "", errors.New("missing prompt action")
	}
	return p.actions[len(p.paths)-1], nil
}

func twentyNumberedLines(prefix string) string {
	return prefix + " 01\n" +
		prefix + " 02\n" +
		prefix + " 03\n" +
		prefix + " 04\n" +
		prefix + " 05\n" +
		prefix + " 06\n" +
		prefix + " 07\n" +
		prefix + " 08\n" +
		prefix + " 09\n" +
		prefix + " 10\n" +
		prefix + " 11\n" +
		prefix + " 12\n" +
		prefix + " 13\n" +
		prefix + " 14\n" +
		prefix + " 15\n" +
		prefix + " 16\n" +
		prefix + " 17\n" +
		prefix + " 18\n" +
		prefix + " 19\n" +
		prefix + " 20"
}

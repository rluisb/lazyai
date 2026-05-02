package wizard

import (
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

package adapter

import (
	"strings"
	"testing"
)

func TestMergeManagedBlock_CreateNew(t *testing.T) {
	managed := []byte("managed content")
	result := MergeManagedBlock(nil, managed, ManagedBlockStartMarker, ManagedBlockEndMarker)
	if !strings.Contains(string(result), ManagedBlockStartMarker) {
		t.Error("missing start marker")
	}
	if !strings.Contains(string(result), ManagedBlockEndMarker) {
		t.Error("missing end marker")
	}
	if !strings.Contains(string(result), "managed content") {
		t.Error("missing managed content")
	}
}

func TestMergeManagedBlock_AppendToExisting(t *testing.T) {
	existing := []byte("# User Content\n")
	managed := []byte("managed content")
	result := MergeManagedBlock(existing, managed, ManagedBlockStartMarker, ManagedBlockEndMarker)
	if !strings.Contains(string(result), "# User Content") {
		t.Error("existing content was lost")
	}
	if !strings.Contains(string(result), ManagedBlockStartMarker) {
		t.Error("missing start marker")
	}
	if !strings.Contains(string(result), ManagedBlockEndMarker) {
		t.Error("missing end marker")
	}
}

func TestMergeManagedBlock_ReplaceExistingBlock(t *testing.T) {
	existing := []byte("# User Content\n\n" + ManagedBlockStartMarker + "\nold content\n" + ManagedBlockEndMarker + "\n\nmore user content")
	managed := []byte("new managed content")
	result := MergeManagedBlock(existing, managed, ManagedBlockStartMarker, ManagedBlockEndMarker)
	if !strings.Contains(string(result), "# User Content") {
		t.Error("pre-block user content was lost")
	}
	if !strings.Contains(string(result), "more user content") {
		t.Error("post-block user content was lost")
	}
	if !strings.Contains(string(result), "new managed content") {
		t.Error("new managed content not inserted")
	}
	if strings.Contains(string(result), "old content") {
		t.Error("old managed content was not replaced")
	}
}

func TestMergeManagedBlock_PayloadContainsEndMarker(t *testing.T) {
	// Regression for #423: if managed payload contains the end-marker literal,
	// the old code matched the wrong end marker and leaked trailing content.
	existing := []byte("# User\n\n" +
		ManagedBlockStartMarker + "\nold\n" + ManagedBlockEndMarker + "\n\n# Tail\n")
	// New payload contains the end-marker literal inside it.
	payload := []byte("line1\n" + ManagedBlockEndMarker + " (escaped ref)\nline3")
	result := string(MergeManagedBlock(existing, payload, ManagedBlockStartMarker, ManagedBlockEndMarker))

	if !strings.Contains(result, "# User") {
		t.Error("pre-block user content was lost")
	}
	if !strings.Contains(result, "# Tail") {
		t.Error("post-block user content was lost")
	}
	if !strings.Contains(result, "line1") || !strings.Contains(result, "line3") {
		t.Error("new managed content not fully inserted")
	}
}

func TestMergeManagedBlock_MultipleBlocks(t *testing.T) {
	// Regression for #423: two independent managed blocks in the same file.
	startA, endA := "<!-- start:A -->", "<!-- end:A -->"
	startB, endB := "<!-- start:B -->", "<!-- end:B -->"
	existing := []byte("preamble\n" +
		startA + "\nblockA\n" + endA + "\nmiddle\n" +
		startB + "\nblockB\n" + endB + "\npostamble\n")

	result := string(MergeManagedBlock(existing, []byte("newA"), startA, endA))
	if !strings.Contains(result, "newA") {
		t.Error("block A content not replaced")
	}
	// Block B must remain intact.
	if !strings.Contains(result, "blockB") {
		t.Error("block B was corrupted")
	}
	if !strings.Contains(result, "postamble") {
		t.Error("postamble lost")
	}
}

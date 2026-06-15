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

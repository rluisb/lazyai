package ledger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenGenesis(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ledger.jsonl")

	l, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	entries, err := l.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 genesis entry, got %d", len(entries))
	}

	genesis := entries[0]
	if genesis.Seq != 1 {
		t.Errorf("genesis seq = %d, want 1", genesis.Seq)
	}
	if genesis.Type != "genesis" {
		t.Errorf("genesis type = %q, want genesis", genesis.Type)
	}
	if genesis.PrevHash != strings.Repeat("0", 64) {
		t.Errorf("genesis prev_hash = %q, want 64 zeros", genesis.PrevHash)
	}
	if genesis.Hash == "" {
		t.Error("genesis hash is empty")
	}
}

func TestAppendAndVerify(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ledger.jsonl")

	l, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	// Append entries
	for i := 0; i < 3; i++ {
		data, _ := json.Marshal(map[string]string{"action": fmt.Sprintf("test-%d", i)})
		entry := &Entry{
			Type:      "test",
			SessionID: "ses_123",
			Data:      data,
		}
		if err := l.Append(entry); err != nil {
			t.Fatalf("Append %d failed: %v", i, err)
		}
	}

	// Verify
	if err := l.Verify(); err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	// Check entries
	entries, _ := l.ReadAll()
	if len(entries) != 4 { // genesis + 3
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	// Check chain
	for i := 1; i < len(entries); i++ {
		if entries[i].PrevHash != entries[i-1].Hash {
			t.Errorf("entry %d: prev_hash != prev entry hash", i)
		}
		if entries[i].Seq != i+1 {
			t.Errorf("entry %d: seq = %d, want %d", i, entries[i].Seq, i+1)
		}
	}
}

func TestVerifyBrokenChain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ledger.jsonl")

	l, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	// Append valid entry
	data, _ := json.Marshal(map[string]string{"action": "valid"})
	entry := &Entry{Type: "test", Data: data}
	if err := l.Append(entry); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Corrupt the file by modifying the hash
	content, _ := os.ReadFile(path)
	corrupted := strings.Replace(string(content), entry.Hash, "badhash000000000000000000000000000000000000000000000000000000000000", 1)
	os.WriteFile(path, []byte(corrupted), 0o644)

	// Re-open and verify
	l2, _ := Open(path)
	if err := l2.Verify(); err == nil {
		t.Fatal("expected Verify to fail with corrupted hash")
	}
}

func TestRedaction(t *testing.T) {
	data, _ := json.Marshal(map[string]interface{}{
		"user":    "alice",
		"token":   "secret123",
		"api_key": "key456",
		"nested": map[string]interface{}{
			"password": "hunter2",
		},
	})

	redacted, redactions := redactSecrets(data)

	var obj map[string]interface{}
	json.Unmarshal(redacted, &obj)

	if obj["token"] != "[REDACTED]" {
		t.Errorf("token = %q, want [REDACTED]", obj["token"])
	}
	if obj["api_key"] != "[REDACTED]" {
		t.Errorf("api_key = %q, want [REDACTED]", obj["api_key"])
	}
	if obj["user"] != "alice" {
		t.Errorf("user = %q, want alice", obj["user"])
	}

	nested := obj["nested"].(map[string]interface{})
	if nested["password"] != "[REDACTED]" {
		t.Errorf("nested.password = %q, want [REDACTED]", nested["password"])
	}

	if len(redactions) != 3 {
		t.Errorf("redactions = %d, want 3", len(redactions))
	}
}

func TestComputeHashConsistency(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ledger.jsonl")
	l, _ := Open(path)

	entry := &Entry{
		Seq:       2,
		Timestamp: "2024-01-15T10:30:00Z",
		Type:      "dispatch",
		SessionID: "ses_abc",
		Data:      json.RawMessage(`{"agent":"wall-builder"}`),
		PrevHash:  strings.Repeat("0", 64),
	}

	hash1 := l.computeHash(entry)
	hash2 := l.computeHash(entry)

	if hash1 != hash2 {
		t.Error("hash computation is not deterministic")
	}
}

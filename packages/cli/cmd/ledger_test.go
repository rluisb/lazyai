package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime/ledger"
	"github.com/spf13/cobra"
)

func TestReadLedgerEntries(t *testing.T) {
	tmpDir := t.TempDir()
	ledgerPath := filepath.Join(tmpDir, "test_ledger.jsonl")

	l, err := ledger.Open(ledgerPath)
	if err != nil {
		t.Fatalf("failed to open ledger: %v", err)
	}

	if err := l.Append(&ledger.Entry{Type: "test", Data: []byte(`{"message":"value1"}`)}); err != nil {
		t.Fatalf("failed to append first entry: %v", err)
	}
	if err := l.Append(&ledger.Entry{Type: "test", Data: []byte(`{"message":"value2"}`)}); err != nil {
		t.Fatalf("failed to append second entry: %v", err)
	}

	readEntries, err := readLedgerEntries(ledgerPath)
	if err != nil {
		t.Fatalf("failed to read ledger: %v", err)
	}

	if len(readEntries) != 3 {
		t.Errorf("expected 3 entries (genesis + 2), got %d", len(readEntries))
	}

	if len(readEntries) > 1 {
		if readEntries[1].EventType != "test" {
			t.Errorf("expected event type 'test', got '%s'", readEntries[1].EventType)
		}
		if !strings.Contains(readEntries[1].Data, "value1") {
			t.Errorf("expected data to contain 'value1', got '%s'", readEntries[1].Data)
		}
	}

	if len(readEntries) > 2 {
		if readEntries[2].EventType != "test" {
			t.Errorf("expected event type 'test', got '%s'", readEntries[2].EventType)
		}
		if !strings.Contains(readEntries[2].Data, "value2") {
			t.Errorf("expected data to contain 'value2', got '%s'", readEntries[2].Data)
		}
		if readEntries[2].PrevHash != readEntries[1].Hash {
			t.Error("hash chain is broken")
		}
	}
}

func TestInitLedger(t *testing.T) {
	tmpDir := t.TempDir()
	ledgerPath := filepath.Join(tmpDir, "test_ledger.jsonl")

	if err := initLedger(ledgerPath); err != nil {
		t.Fatalf("failed to initialize ledger: %v", err)
	}

	if _, err := os.Stat(ledgerPath); os.IsNotExist(err) {
		t.Error("ledger file was not created")
	}

	l, err := ledger.Open(ledgerPath)
	if err != nil {
		t.Fatalf("failed to open ledger: %v", err)
	}

	entries, err := l.ReadAll()
	if err != nil {
		t.Fatalf("failed to read ledger: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 genesis entry, got %d", len(entries))
	}

	if len(entries) > 0 {
		if entries[0].Type != "genesis" {
			t.Errorf("expected genesis entry, got '%s'", entries[0].Type)
		}
		if entries[0].PrevHash != strings.Repeat("0", 64) {
			t.Errorf("expected genesis prev_hash to be 64 zeros, got '%s'", entries[0].PrevHash)
		}
	}
}

func TestAppendToLedger(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp directory: %v", err)
	}
	defer os.Chdir(origDir)

	ledgerPath := getLedgerPath()

	if err := initLedger(ledgerPath); err != nil {
		t.Fatalf("failed to initialize ledger: %v", err)
	}

	if err := appendToLedger("test_event", map[string]string{"message": "value"}); err != nil {
		t.Fatalf("failed to append to ledger: %v", err)
	}

	l, err := ledger.Open(ledgerPath)
	if err != nil {
		t.Fatalf("failed to open ledger: %v", err)
	}

	entries, err := l.ReadAll()
	if err != nil {
		t.Fatalf("failed to read ledger: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	if len(entries) > 1 {
		if entries[1].Type != "test_event" {
			t.Errorf("expected event type 'test_event', got '%s'", entries[1].Type)
		}

		var dataMap map[string]string
		if err := json.Unmarshal(entries[1].Data, &dataMap); err != nil {
			t.Fatalf("failed to unmarshal entry data: %v", err)
		}
		if dataMap["message"] != "value" {
			t.Errorf("expected data message=value, got '%s'", dataMap["message"])
		}

		if entries[1].PrevHash != entries[0].Hash {
			t.Error("hash chain is broken")
		}
	}
}

// ── Integration tests for runtime-backed ledger CLI commands ─────────────────

func TestLedgerInitRuntime(t *testing.T) {
	tmpDir := withTempDir(t)

	out := captureStdout(t, func() {
		if err := runLedgerInit(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runLedgerInit failed: %v", err)
		}
	})

	if !strings.Contains(out, "Ledger initialized") {
		t.Errorf("expected output to contain 'Ledger initialized', got:\n%s", out)
	}

	ledgerPath := filepath.Join(tmpDir, ".specify", "ledger.jsonl")
	if _, err := os.Stat(ledgerPath); os.IsNotExist(err) {
		t.Errorf("expected ledger file to exist at %s", ledgerPath)
	}

	entries, err := readLedger(ledgerPath)
	if err != nil {
		t.Fatalf("failed to read ledger: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 genesis entry, got %d", len(entries))
	}
	if len(entries) > 0 {
		if entries[0].EventType != "genesis" {
			t.Errorf("expected genesis entry, got '%s'", entries[0].EventType)
		}
	}
}

func TestLedgerAppendRuntime(t *testing.T) {
	tmpDir := withTempDir(t)
	ledgerPath := filepath.Join(tmpDir, ".specify", "ledger.jsonl")

	if err := initLedger(ledgerPath); err != nil {
		t.Fatalf("failed to init ledger: %v", err)
	}

	out := captureStdout(t, func() {
		if err := runLedgerAppend(&cobra.Command{}, []string{"test_event", "test data"}); err != nil {
			t.Fatalf("runLedgerAppend failed: %v", err)
		}
	})

	if !strings.Contains(out, "Entry appended") {
		t.Errorf("expected output to contain 'Entry appended', got:\n%s", out)
	}

	entries, err := readLedger(ledgerPath)
	if err != nil {
		t.Fatalf("failed to read ledger: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestLedgerShowRuntime(t *testing.T) {
	tmpDir := withTempDir(t)
	ledgerPath := filepath.Join(tmpDir, ".specify", "ledger.jsonl")

	if err := initLedger(ledgerPath); err != nil {
		t.Fatalf("failed to init ledger: %v", err)
	}
	if err := appendToLedger("show_test", map[string]string{"data": "show me"}); err != nil {
		t.Fatalf("failed to append to ledger: %v", err)
	}

	out := captureStdout(t, func() {
		if err := runLedgerShow(&cobra.Command{}, []string{"5"}); err != nil {
			t.Fatalf("runLedgerShow failed: %v", err)
		}
	})

	if !strings.Contains(out, "show_test") {
		t.Errorf("expected output to contain 'show_test', got:\n%s", out)
	}
}

func TestLedgerVerifyRuntime(t *testing.T) {
	tmpDir := withTempDir(t)
	ledgerPath := filepath.Join(tmpDir, ".specify", "ledger.jsonl")

	if err := initLedger(ledgerPath); err != nil {
		t.Fatalf("failed to init ledger: %v", err)
	}
	if err := appendToLedger("verify_test", map[string]string{"data": "test"}); err != nil {
		t.Fatalf("failed to append to ledger: %v", err)
	}

	out := captureStdout(t, func() {
		if err := runLedgerVerify(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runLedgerVerify failed: %v", err)
		}
	})

	if !strings.Contains(out, "verified") {
		t.Errorf("expected output to contain 'verified', got:\n%s", out)
	}
}

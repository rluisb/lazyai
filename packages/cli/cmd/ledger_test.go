package cmd

import (
			"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSha256Hash(t *testing.T) {
	// Test that sha256Hash produces consistent results
	data := []byte("test data")
	hash1 := sha256Hash(data)
	hash2 := sha256Hash(data)
	
	if hash1 != hash2 {
		t.Error("sha256Hash is not deterministic")
	}
	
	// Should be 64 characters (hex encoded SHA-256)
	if len(hash1) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}
	
	// Different data should produce different hashes
	differentData := []byte("different data")
	differentHash := sha256Hash(differentData)
	
	if hash1 == differentHash {
		t.Error("Different data produced same hash")
	}
}

func TestLedgerEntryHash(t *testing.T) {
	entry := LedgerEntry{
		ID:        "test_entry",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventType: "test",
		Data:      "test data",
		PrevHash:  "",
	}
	
	entryData, _ := json.Marshal(map[string]string{
		"id":        entry.ID,
		"timestamp": entry.Timestamp,
		"event_type": entry.EventType,
		"data":      entry.Data,
		"prev_hash": entry.PrevHash,
	})
	
	entry.Hash = sha256Hash(entryData)
	
	// Verify hash is not empty
	if entry.Hash == "" {
		t.Error("Hash should not be empty")
	}
	
	// Verify hash length
	if len(entry.Hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(entry.Hash))
	}
}

func TestReadLedgerEntries(t *testing.T) {
	// Create a temporary ledger file
	tmpDir := t.TempDir()
	ledgerPath := filepath.Join(tmpDir, "test_ledger.jsonl")
	
	// Create test entries
	entries := []LedgerEntry{
		{
			ID:        "entry1",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			EventType: "test",
			Data:      "test data 1",
			Hash:      "hash1",
			PrevHash:  "",
		},
		{
			ID:        "entry2",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			EventType: "test",
			Data:      "test data 2",
			Hash:      "hash2",
			PrevHash:  "hash1",
		},
	}
	
	// Write entries to file
	file, err := os.Create(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create test ledger: %v", err)
	}
	
	for _, entry := range entries {
		entryJSON, _ := json.Marshal(entry)
		file.WriteString(string(entryJSON) + "\n")
	}
	file.Close()
	
	// Read entries back
	readEntries, err := readLedgerEntries(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to read ledger: %v", err)
	}
	
	if len(readEntries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(readEntries))
	}
	
	// Verify first entry
	if len(readEntries) > 0 {
		if readEntries[0].ID != "entry1" {
			t.Errorf("Expected ID 'entry1', got '%s'", readEntries[0].ID)
		}
	}
	
	// Verify second entry
	if len(readEntries) > 1 {
		if readEntries[1].ID != "entry2" {
			t.Errorf("Expected ID 'entry2', got '%s'", readEntries[1].ID)
		}
		if readEntries[1].PrevHash != "hash1" {
			t.Errorf("Expected PrevHash 'hash1', got '%s'", readEntries[1].PrevHash)
		}
	}
}

func TestInitLedger(t *testing.T) {
	tmpDir := t.TempDir()
	ledgerPath := filepath.Join(tmpDir, "test_ledger.jsonl")
	
	err := initLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to initialize ledger: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(ledgerPath); os.IsNotExist(err) {
		t.Error("Ledger file was not created")
	}
	
	// Read entries
	entries, err := readLedgerEntries(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to read ledger: %v", err)
	}
	
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	
	if len(entries) > 0 {
		if entries[0].EventType != "genesis" {
			t.Errorf("Expected event type 'genesis', got '%s'", entries[0].EventType)
		}
	}
}

func TestAppendToLedger(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and switch to temp directory because appendToLedger uses os.Getwd()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir to temp directory: %v", err)
	}
	defer os.Chdir(origDir)

	ledgerPath := getLedgerPath()

	// Initialize ledger
	err = initLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to initialize ledger: %v", err)
	}

	// Append an event
	err = appendToLedger("test_event", map[string]string{
		"key": "value",
	})
	if err != nil {
		t.Fatalf("Failed to append to ledger: %v", err)
	}

	// Read entries
	entries, err := readLedgerEntries(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to read ledger: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	if len(entries) > 1 {
		if entries[1].EventType != "test_event" {
			t.Errorf("Expected event type 'test_event', got '%s'", entries[1].EventType)
		}

		if !strings.Contains(entries[1].Data, "key=value") {
			t.Errorf("Expected data to contain 'key=value', got '%s'", entries[1].Data)
		}

		// Verify hash chain
		if entries[1].PrevHash != entries[0].Hash {
			t.Error("Hash chain is broken")
		}
	}
}

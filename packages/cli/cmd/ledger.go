package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var ledgerCmd = &cobra.Command{
	Use:   "ledger",
	Short: "Immutable audit trail",
	Long:  "Append-only hash-chained ledger for tracking agent actions and decisions.",
}

var ledgerInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ledger",
	RunE:  runLedgerInit,
}

var ledgerAppendCmd = &cobra.Command{
	Use:   "append [event-type] [data]",
	Short: "Append event to ledger",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runLedgerAppend,
}

var ledgerVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify ledger integrity",
	RunE:  runLedgerVerify,
}

var ledgerShowCmd = &cobra.Command{
	Use:   "show [n]",
	Short: "Show last n entries",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLedgerShow,
}

func init() {
	ledgerCmd.AddCommand(ledgerInitCmd)
	ledgerCmd.AddCommand(ledgerAppendCmd)
	ledgerCmd.AddCommand(ledgerVerifyCmd)
	ledgerCmd.AddCommand(ledgerShowCmd)
	rootCmd.AddCommand(ledgerCmd)
	ledgerCmd.GroupID = "audit"
}

// LedgerEntry represents a single ledger entry
type LedgerEntry struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	EventType string `json:"event_type"`
	Data      string `json:"data"`
	Hash      string `json:"hash"`
	PrevHash  string `json:"prev_hash"`
}

func getLedgerPath() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, ".specify", "ledger.jsonl")
}

func runLedgerInit(cmd *cobra.Command, args []string) error {
	ledgerPath := getLedgerPath()

	if _, err := os.Stat(ledgerPath); err == nil {
		return fmt.Errorf("ledger already exists at %s", ledgerPath)
	}

	dir := filepath.Dir(ledgerPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	genesis := LedgerEntry{
		ID:        "genesis",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventType: "genesis",
		Data:      "Ledger initialized",
		Hash:      "",
		PrevHash:  "",
	}

	genesisData, _ := json.Marshal(map[string]string{
		"id":         genesis.ID,
		"timestamp":  genesis.Timestamp,
		"event_type": genesis.EventType,
		"data":       genesis.Data,
		"prev_hash":  genesis.PrevHash,
	})
	genesis.Hash = sha256Hash(genesisData)

	file, err := os.Create(ledgerPath)
	if err != nil {
		return fmt.Errorf("failed to create ledger: %w", err)
	}
	defer file.Close()

	entryJSON, _ := json.Marshal(genesis)
	file.WriteString(string(entryJSON) + "\n")

	fmt.Printf("✅ Ledger initialized at %s\n", ledgerPath)
	fmt.Printf("   Genesis hash: %s\n", genesis.Hash)
	return nil
}

func runLedgerAppend(cmd *cobra.Command, args []string) error {
	ledgerPath := getLedgerPath()

	entries, err := readLedger(ledgerPath)
	if err != nil {
		return err
	}

	var prevHash string
	if len(entries) > 0 {
		prevHash = entries[len(entries)-1].Hash
	}

	eventType := args[0]
	data := ""
	if len(args) > 1 {
		data = strings.Join(args[1:], " ")
	}

	entry := LedgerEntry{
		ID:        fmt.Sprintf("entry_%d", time.Now().UnixNano()),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventType: eventType,
		Data:      data,
		PrevHash:  prevHash,
	}

	entryData, _ := json.Marshal(map[string]string{
		"id":         entry.ID,
		"timestamp":  entry.Timestamp,
		"event_type": entry.EventType,
		"data":       entry.Data,
		"prev_hash":  entry.PrevHash,
	})
	entry.Hash = sha256Hash(entryData)

	file, err := os.OpenFile(ledgerPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open ledger: %w", err)
	}
	defer file.Close()

	entryJSON, _ := json.Marshal(entry)
	file.WriteString(string(entryJSON) + "\n")

	fmt.Printf("✅ Entry appended\n")
	fmt.Printf("   Type: %s\n", eventType)
	fmt.Printf("   Hash: %s\n", entry.Hash)
	return nil
}

func runLedgerVerify(cmd *cobra.Command, args []string) error {
	ledgerPath := getLedgerPath()
	entries, err := readLedger(ledgerPath)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return fmt.Errorf("ledger is empty")
	}

	fmt.Println("🔍 Verifying ledger integrity...")
	fmt.Println()

	verified := 0
	broken := 0

	for i, entry := range entries {
		if i > 0 {
			if entry.PrevHash != entries[i-1].Hash {
				fmt.Printf("  ❌ Entry %d: hash chain broken\n", i)
				fmt.Printf("     Expected prev_hash: %s\n", entries[i-1].Hash)
				fmt.Printf("     Actual prev_hash:   %s\n", entry.PrevHash)
				broken++
				continue
			}
		}

		entryData, _ := json.Marshal(map[string]string{
			"id":         entry.ID,
			"timestamp":  entry.Timestamp,
			"event_type": entry.EventType,
			"data":       entry.Data,
			"prev_hash":  entry.PrevHash,
		})
		expectedHash := sha256Hash(entryData)

		if entry.Hash != expectedHash {
			fmt.Printf("  ❌ Entry %d: hash mismatch\n", i)
			fmt.Printf("     Expected: %s\n", expectedHash)
			fmt.Printf("     Actual:   %s\n", entry.Hash)
			broken++
			continue
		}

		verified++
	}

	fmt.Println()
	if broken > 0 {
		fmt.Printf("  ❌ Ledger verification FAILED: %d/%d entries broken\n", broken, len(entries))
		return fmt.Errorf("ledger verification failed")
	}

	fmt.Printf("  ✅ All %d entries verified. Chain intact.\n", len(entries))
	return nil
}

func runLedgerShow(cmd *cobra.Command, args []string) error {
	ledgerPath := getLedgerPath()
	entries, err := readLedger(ledgerPath)
	if err != nil {
		return err
	}

	n := 10
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &n)
	}

	if n > len(entries) {
		n = len(entries)
	}

	start := len(entries) - n
	if start < 0 {
		start = 0
	}

	fmt.Printf("Last %d entries:\n", n)
	fmt.Println(strings.Repeat("-", 80))

	for i := start; i < len(entries); i++ {
		entry := entries[i]
		fmt.Printf("[%s] %s | %s\n", entry.Timestamp, entry.EventType, entry.Data)
		fmt.Printf("  Hash: %s...%s\n", entry.Hash[:8], entry.Hash[len(entry.Hash)-8:])
		if entry.PrevHash != "" {
			fmt.Printf("  Prev: %s...%s\n", entry.PrevHash[:8], entry.PrevHash[len(entry.PrevHash)-8:])
		}
		fmt.Println()
	}

	return nil
}

func readLedger(path string) ([]LedgerEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ledger: %w", err)
	}

	var entries []LedgerEntry
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry LedgerEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func sha256Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

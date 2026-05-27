package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime/ledger"
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

// LedgerEntry represents a single ledger entry (legacy display format)
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

	l, err := ledger.Open(ledgerPath)
	if err != nil {
		return fmt.Errorf("failed to initialize ledger: %w", err)
	}
	_ = l // ledger.Open already creates genesis

	fmt.Printf("✅ Ledger initialized at %s\n", ledgerPath)
	fmt.Printf("   Genesis hash: %s\n", strings.Repeat("0", 64))
	return nil
}

func runLedgerAppend(cmd *cobra.Command, args []string) error {
	ledgerPath := getLedgerPath()

	l, err := ledger.Open(ledgerPath)
	if err != nil {
		return fmt.Errorf("failed to open ledger: %w", err)
	}

	eventType := args[0]
	data := ""
	if len(args) > 1 {
		data = strings.Join(args[1:], " ")
	}

	// Build JSON data payload
	dataJSON, _ := json.Marshal(map[string]string{"data": data})

	entry := &ledger.Entry{
		Type: eventType,
		Data: dataJSON,
	}

	if err := l.Append(entry); err != nil {
		return fmt.Errorf("failed to append to ledger: %w", err)
	}

	fmt.Printf("✅ Entry appended\n")
	fmt.Printf("   Type: %s\n", eventType)
	fmt.Printf("   Hash: %s\n", entry.Hash)
	return nil
}

func runLedgerVerify(cmd *cobra.Command, args []string) error {
	ledgerPath := getLedgerPath()

	l, err := ledger.Open(ledgerPath)
	if err != nil {
		return fmt.Errorf("failed to open ledger: %w", err)
	}

	fmt.Println("🔍 Verifying ledger integrity...")
	fmt.Println()

	if err := l.Verify(); err != nil {
		fmt.Printf("  ❌ Ledger verification FAILED: %v\n", err)
		return fmt.Errorf("ledger verification failed")
	}

	entries, _ := l.ReadAll()
	fmt.Printf("  ✅ All %d entries verified. Chain intact.\n", len(entries))
	return nil
}

func runLedgerShow(cmd *cobra.Command, args []string) error {
	ledgerPath := getLedgerPath()

	l, err := ledger.Open(ledgerPath)
	if err != nil {
		return fmt.Errorf("failed to open ledger: %w", err)
	}

	entries, err := l.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read ledger: %w", err)
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
		// Extract data string for display
		dataStr := string(entry.Data)
		var dataMap map[string]string
		if err := json.Unmarshal(entry.Data, &dataMap); err == nil {
			if d, ok := dataMap["data"]; ok {
				dataStr = d
			}
		}
		fmt.Printf("[%s] %s | %s\n", entry.Timestamp, entry.Type, dataStr)
		fmt.Printf("  Hash: %s...%s\n", entry.Hash[:8], entry.Hash[len(entry.Hash)-8:])
		if entry.PrevHash != "" {
			fmt.Printf("  Prev: %s...%s\n", entry.PrevHash[:8], entry.PrevHash[len(entry.PrevHash)-8:])
		}
		fmt.Println()
	}

	return nil
}

func readLedger(path string) ([]LedgerEntry, error) {
	l, err := ledger.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open ledger: %w", err)
	}

	entries, err := l.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read ledger: %w", err)
	}

	var result []LedgerEntry
	for _, e := range entries {
		dataStr := string(e.Data)
		var dataMap map[string]string
		if err := json.Unmarshal(e.Data, &dataMap); err == nil {
			if d, ok := dataMap["data"]; ok {
				dataStr = d
			}
		}
		result = append(result, LedgerEntry{
			ID:        fmt.Sprintf("entry_%d", e.Seq),
			Timestamp: e.Timestamp,
			EventType: e.Type,
			Data:      dataStr,
			Hash:      e.Hash,
			PrevHash:  e.PrevHash,
		})
	}

	return result, nil
}

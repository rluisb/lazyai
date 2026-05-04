package main

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFindRunningServerDoesNotClearDiscoveryOnHealthTimeout(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second)
		w.WriteHeader(http.StatusOK)
	})}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	go server.Serve(listener)

	info := daemonInfo{Port: listener.Addr().(*net.TCPAddr).Port, PID: os.Getpid(), StartedAt: time.Now().UTC().Format(time.RFC3339)}
	b, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(discoveryPath()), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(discoveryPath(), b, 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := findRunningServer(); err == nil {
		t.Fatal("expected health timeout error")
	}
	if _, err := os.Stat(discoveryPath()); err != nil {
		t.Fatalf("expected discovery to remain after timeout: %v", err)
	}
}

func TestStartDetachedRejectsInvalidExecutionMode(t *testing.T) {
	err := startDetachedWithStatus(0, t.TempDir(), "project", "bogus", "", true, false, 0, nil)
	if err == nil {
		t.Fatalf("expected invalid execution mode error")
	}
	if !strings.Contains(err.Error(), "invalid execution mode") {
		t.Fatalf("unexpected error for invalid execution mode: %v", err)
	}
}

package dashboard

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/catalog"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestDashboardEventsJSONReplayUsesSinceIDAndValidatesRunKind(t *testing.T) {
	database := newDashboardTestDB(t)
	seedRun(t, database, types.RunKindChain, "chain-events", "release", "1", "running", "build", chainStateJSON(t, "chain-events", "release", "1", "running", "build"), "2026-05-05T10:00:00Z")
	seedEvent(t, database, "chain-events", "step_started", `{"stepId":"build"}`, "2026-05-05T10:01:00Z")
	seedEvent(t, database, "chain-events", "step_completed", `{"stepId":"build"}`, "2026-05-05T10:02:00Z")
	seedEvent(t, database, "other-run", "other_event", `{}`, "2026-05-05T10:03:00Z")
	handler := newDashboardHTTPHandler(t, database, catalog.NewStore(database))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/dashboard/runs/chain/chain-events/events?since_id=1", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("events replay status = %d body=%s", response.Code, response.Body.String())
	}
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	var replay RunEventsResponse
	decodeResponse(t, response, &replay)
	if len(replay.Items) != 1 || replay.Items[0].ID != 2 || replay.Items[0].EventType != "step_completed" {
		t.Fatalf("replay mismatch: %+v", replay)
	}

	badSinceResponse := httptest.NewRecorder()
	handler.ServeHTTP(badSinceResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/runs/chain/chain-events/events?since_id=bad", nil))
	if badSinceResponse.Code != http.StatusBadRequest {
		t.Fatalf("bad since_id status = %d body=%s", badSinceResponse.Code, badSinceResponse.Body.String())
	}

	wrongKindResponse := httptest.NewRecorder()
	handler.ServeHTTP(wrongKindResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/runs/team/chain-events/events", nil))
	if wrongKindResponse.Code != http.StatusNotFound {
		t.Fatalf("wrong kind status = %d body=%s", wrongKindResponse.Code, wrongKindResponse.Body.String())
	}
}

func TestDashboardEventsSSEReplaysAndStreamsLivePerRunOnly(t *testing.T) {
	database := newDashboardTestDB(t)
	seedRun(t, database, types.RunKindChain, "chain-stream", "release", "1", "running", "build", chainStateJSON(t, "chain-stream", "release", "1", "running", "build"), "2026-05-05T10:00:00Z")
	seedRun(t, database, types.RunKindChain, "chain-other", "release", "1", "running", "test", chainStateJSON(t, "chain-other", "release", "1", "running", "test"), "2026-05-05T10:00:00Z")
	seedEvent(t, database, "chain-stream", "step_started", `{"stepId":"build"}`, "2026-05-05T10:01:00Z")
	handler, bus := newDashboardHTTPHandlerWithBus(t, database, catalog.NewStore(database))
	server := httptest.NewServer(handler)
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/dashboard/runs/chain/chain-stream/events?since_id=0", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	request.Header.Set("Accept", "text/event-stream")
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatalf("open sse stream: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("sse status = %d", response.StatusCode)
	}
	if !strings.HasPrefix(response.Header.Get("Content-Type"), "text/event-stream") {
		t.Fatalf("sse content type = %q", response.Header.Get("Content-Type"))
	}

	reader := bufio.NewReader(response.Body)
	if !readUntil(t, reader, "step_started", time.Second) {
		t.Fatal("did not receive replayed event")
	}

	go bus.Publish("chain-stream", "step_completed", map[string]any{"stepId": "build"})
	if !readUntil(t, reader, "step_completed", time.Second) {
		t.Fatal("did not receive live event")
	}

	bus.Publish("chain-other", "other_event", map[string]any{"stepId": "test"})
	if readUntil(t, reader, "other_event", 100*time.Millisecond) {
		t.Fatal("received event for a different run")
	}
}

func readUntil(t *testing.T, reader *bufio.Reader, contains string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.After(timeout)
	lines := make(chan string, 1)
	errs := make(chan error, 1)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				errs <- err
				return
			}
			if strings.Contains(line, contains) {
				lines <- line
				return
			}
		}
	}()
	select {
	case <-lines:
		return true
	case <-errs:
		return false
	case <-deadline:
		return false
	}
}

func TestSSEEventEncodingIsStableJSON(t *testing.T) {
	event := DashboardEvent{ID: 7, RunID: "chain-1", EventType: "step_started", Data: map[string]any{"stepId": "build"}, CreatedAt: "2026-05-05T10:00:00Z"}
	encoded, err := formatSSEEvent(event)
	if err != nil {
		t.Fatalf("format event: %v", err)
	}
	if !strings.Contains(encoded, "id: 7\n") || !strings.Contains(encoded, "event: step_started\n") || !strings.HasSuffix(encoded, "\n\n") {
		t.Fatalf("sse framing mismatch: %q", encoded)
	}
	var decoded DashboardEvent
	dataLine := ""
	for _, line := range strings.Split(encoded, "\n") {
		if strings.HasPrefix(line, "data: ") {
			dataLine = strings.TrimPrefix(line, "data: ")
		}
	}
	if err := json.Unmarshal([]byte(dataLine), &decoded); err != nil {
		t.Fatalf("decode data line %q: %v", dataLine, err)
	}
	if decoded.ID != event.ID || decoded.EventType != event.EventType || decoded.Data["stepId"] != "build" {
		t.Fatalf("decoded event mismatch: %+v", decoded)
	}
}

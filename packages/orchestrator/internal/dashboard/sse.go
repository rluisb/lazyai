package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/events"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

// handleGlobalEvents serves /api/dashboard/events with optional since_id replay
// followed (over SSE) by a live stream of events across all runs. Without an
// SSE-accepting client it returns a bounded JSON snapshot.
func (h *Handler) handleGlobalEvents(w http.ResponseWriter, r *http.Request) {
	if h.events == nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "dashboard event bus is not configured")
		return
	}
	sinceID, err := parseOptionalNonNegativeInt(r.URL.Query().Get("since_id"), "since_id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	limit, err := parseOptionalNonNegativeInt(r.URL.Query().Get("limit"), "limit")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	limit = NormalizeLimit(limit, DefaultEventLimit, MaxEventLimit)

	if acceptsSSE(r.Header.Get("Accept")) {
		h.streamGlobalEvents(w, r, sinceID, limit)
		return
	}
	replay := busEventsToDashboard(h.events.ReplayAll(sinceID, limit))
	h.writeJSON(w, http.StatusOK, RunEventsResponse{Items: replay})
}

func (h *Handler) streamGlobalEvents(w http.ResponseWriter, r *http.Request, sinceID, limit int) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "streaming is not supported by this response writer")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	ch := make(chan events.Event, 64)
	h.events.SubscribeAll(ch)
	defer h.events.UnsubscribeAll(ch)

	// Replay first so the client can reconcile against persisted state.
	lastReplayedID := sinceID
	for _, event := range busEventsToDashboard(h.events.ReplayAll(sinceID, limit)) {
		if err := writeSSEEvent(w, event); err != nil {
			return
		}
		if event.ID > lastReplayedID {
			lastReplayedID = event.ID
		}
		flusher.Flush()
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-ch:
			converted := dashboardEventFromBus(event)
			// Skip events that the replay already delivered. New live events have
			// id 0 in the bus (the bus does not round-trip through the DB before
			// fanout), so we always forward those.
			if converted.ID > 0 && converted.ID <= lastReplayedID {
				continue
			}
			if err := writeSSEEvent(w, converted); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (h *Handler) handleRunEvents(w http.ResponseWriter, r *http.Request, kind types.RunKind, id string) {
	if h.readModel == nil || h.events == nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "dashboard event dependencies are not configured")
		return
	}
	if _, err := h.readModel.getRunRow(r.Context(), kind, id); err != nil {
		h.handleError(w, err)
		return
	}
	sinceID, err := parseOptionalNonNegativeInt(r.URL.Query().Get("since_id"), "since_id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	limit, err := parseOptionalNonNegativeInt(r.URL.Query().Get("limit"), "limit")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	limit = NormalizeLimit(limit, DefaultEventLimit, MaxEventLimit)

	if acceptsSSE(r.Header.Get("Accept")) {
		h.streamRunEvents(w, r, id, sinceID, limit)
		return
	}
	events := busEventsToDashboard(h.events.Replay(id, sinceID))
	if len(events) > limit {
		events = events[:limit]
	}
	h.writeJSON(w, http.StatusOK, RunEventsResponse{Items: events})
}

func (h *Handler) streamRunEvents(w http.ResponseWriter, r *http.Request, runID string, sinceID int, limit int) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "streaming is not supported by this response writer")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	ch := make(chan events.Event, 64)
	h.events.Subscribe(runID, ch)
	defer h.events.Unsubscribe(runID, ch)

	replay := busEventsToDashboard(h.events.Replay(runID, sinceID))
	if len(replay) > limit {
		replay = replay[:limit]
	}
	for _, event := range replay {
		if err := writeSSEEvent(w, event); err != nil {
			return
		}
		flusher.Flush()
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-ch:
			if event.RunID != runID {
				continue
			}
			if err := writeSSEEvent(w, dashboardEventFromBus(event)); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func acceptsSSE(accept string) bool {
	for _, part := range strings.Split(accept, ",") {
		mediaType := strings.TrimSpace(strings.Split(part, ";")[0])
		if mediaType == "text/event-stream" {
			return true
		}
	}
	return false
}

func writeSSEEvent(w http.ResponseWriter, event DashboardEvent) error {
	encoded, err := formatSSEEvent(event)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(encoded))
	return err
}

func formatSSEEvent(event DashboardEvent) (string, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	if event.ID > 0 {
		fmt.Fprintf(&builder, "id: %d\n", event.ID)
	}
	if event.EventType != "" {
		fmt.Fprintf(&builder, "event: %s\n", event.EventType)
	}
	builder.WriteString("data: ")
	builder.Write(data)
	builder.WriteString("\n\n")
	return builder.String(), nil
}

func busEventsToDashboard(items []events.Event) []DashboardEvent {
	converted := make([]DashboardEvent, 0, len(items))
	for _, item := range items {
		converted = append(converted, dashboardEventFromBus(item))
	}
	return converted
}

func dashboardEventFromBus(item events.Event) DashboardEvent {
	data := item.Data
	if data == nil {
		data = map[string]any{}
	}
	return DashboardEvent{
		ID:        item.ID,
		RunID:     item.RunID,
		EventType: item.Type,
		Data:      data,
		CreatedAt: item.CreatedAt,
	}
}

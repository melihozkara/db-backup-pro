package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SSEHub manages Server-Sent Events clients
type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan SSEEvent]struct{}
}

// SSEEvent represents an event to be sent to clients
type SSEEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// NewSSEHub creates a new SSE hub
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[chan SSEEvent]struct{}),
	}
}

// Broadcast sends an event to all connected clients
func (h *SSEHub) Broadcast(eventName string, data map[string]interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	event := SSEEvent{
		Event: eventName,
		Data:  data,
	}

	for ch := range h.clients {
		select {
		case ch <- event:
		default:
			// Client buffer full, skip
		}
	}
}

// addClient registers a new SSE client
func (h *SSEHub) addClient() chan SSEEvent {
	ch := make(chan SSEEvent, 64)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

// removeClient unregisters an SSE client
func (h *SSEHub) removeClient(ch chan SSEEvent) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}

// ServeHTTP handles the SSE endpoint
func (h *SSEHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := h.addClient()
	defer h.removeClient(ch)

	// Send initial connected event
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
	flusher.Flush()

	// Heartbeat ticker (30 seconds)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case event := <-ch:
			data, err := json.Marshal(event.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, data)
			flusher.Flush()
		}
	}
}

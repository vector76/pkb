package server

import (
	"encoding/json"
	"sync"
)

// Event is pushed to browser clients over SSE.
type Event struct {
	Type        string `json:"type"`         // "conversation_updated", "wiki_updated", "issues_updated", or "raymond_status"
	Path        string `json:"path"`         // relative KB path, e.g. "conversations/foo.md"
	Active      bool   `json:"active"`       // used by raymond_status events
	UnseenCount int    `json:"unseen_count"` // used by issues_updated events
}

// Hub manages SSE subscribers and fans events out to all of them.
type Hub struct {
	mu      sync.Mutex
	clients map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[chan []byte]struct{})}
}

// Subscribe returns a channel that receives encoded SSE data lines and an
// unsubscribe function the caller must invoke when done.
func (h *Hub) Subscribe() (chan []byte, func()) {
	ch := make(chan []byte, 16)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	unsub := func() {
		h.mu.Lock()
		delete(h.clients, ch)
		h.mu.Unlock()
	}
	return ch, unsub
}

// Publish sends an event to all current subscribers.
// Clients that are not keeping up (full channel) are skipped.
func (h *Hub) Publish(e Event) {
	data, err := json.Marshal(e)
	if err != nil {
		return
	}
	line := append([]byte("data: "), data...)
	line = append(line, '\n', '\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- line:
		default:
			// slow client; skip this event
		}
	}
}

package server

import (
	"context"
	"time"

	"github.com/vector76/pkb/internal/kb"
)

// RaymondPoller periodically checks Raymond liveness and publishes SSE events
// when the active state changes.
type RaymondPoller struct {
	hub *Hub
	kb  *kb.KB
}

func NewRaymondPoller(hub *Hub, kb *kb.KB) *RaymondPoller {
	return &RaymondPoller{hub: hub, kb: kb}
}

// Start polls Raymond liveness every 10 seconds until ctx is cancelled.
func (p *RaymondPoller) Start(ctx context.Context) {
	lastActive := kb.RaymondActive(p.kb, staleThreshold)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			newActive := kb.RaymondActive(p.kb, staleThreshold)
			if newActive != lastActive {
				p.hub.Publish(Event{Type: "raymond_status", Active: newActive})
				lastActive = newActive
			}
		}
	}
}

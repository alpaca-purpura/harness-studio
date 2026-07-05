// Package sse is a multiplexed Server-Sent Events broker built on the standard
// library. A single connection carries typed events (event: map | dock | run);
// reconnecting clients replay missed events via the Last-Event-ID header. It is the
// emitter behind GET /events; the AG-UI taxonomy rides on top of it.
package sse

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
)

// Event types multiplexed on the single stream.
const (
	EventMap  = "map"  // graph/index deltas for the Map (S2).
	EventDock = "dock" // conductor stream-json for the Dock (S4).
	EventRun  = "run"  // run lifecycle for Corridas (S8).
)

// maxHistory caps the replay ring buffer (Last-Event-ID reconnection window).
const maxHistory = 256

// Event is one SSE message. ID is a monotonic sequence; Type is one of the Event*
// constants; Data is the payload (may be multi-line).
type Event struct {
	ID   string
	Type string
	Data []byte
}

// Broker fans events out to every connected subscriber. Safe for concurrent use.
type Broker struct {
	mu      sync.RWMutex
	subs    map[chan Event]struct{}
	history []Event
	seq     uint64
}

// NewBroker returns an empty broker.
func NewBroker() *Broker {
	return &Broker{subs: map[chan Event]struct{}{}}
}

// Publish assigns the next id, records the event for replay, and fans it out. It
// returns the stored event (with its assigned id). Slow subscribers are skipped
// rather than blocking the publisher.
func (b *Broker) Publish(eventType string, data []byte) Event {
	b.mu.Lock()
	b.seq++
	e := Event{ID: strconv.FormatUint(b.seq, 10), Type: eventType, Data: data}
	b.history = append(b.history, e)
	if len(b.history) > maxHistory {
		b.history = b.history[len(b.history)-maxHistory:]
	}
	subs := make([]chan Event, 0, len(b.subs))
	for ch := range b.subs {
		subs = append(subs, ch)
	}
	b.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- e:
		default: // subscriber too slow — it will catch up via Last-Event-ID on reconnect.
		}
	}
	return e
}

// ServeHTTP streams events to one client as text/event-stream, replaying anything
// after Last-Event-ID first, then live events until the request context is done.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")

	sub := b.subscribe()
	defer b.unsubscribe(sub)

	if last := r.Header.Get("Last-Event-ID"); last != "" {
		for _, e := range b.replay(last) {
			writeEvent(w, e)
		}
	}
	fmt.Fprint(w, ": connected\n\n") // open the stream
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case e, open := <-sub:
			if !open {
				return
			}
			writeEvent(w, e)
			flusher.Flush()
		}
	}
}

func (b *Broker) subscribe() chan Event {
	ch := make(chan Event, 16)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Broker) unsubscribe(ch chan Event) {
	b.mu.Lock()
	delete(b.subs, ch)
	b.mu.Unlock()
}

// replay returns the buffered events with an id greater than lastID.
func (b *Broker) replay(lastID string) []Event {
	last, err := strconv.ParseUint(lastID, 10, 64)
	if err != nil {
		return nil
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	var out []Event
	for _, e := range b.history {
		if id, perr := strconv.ParseUint(e.ID, 10, 64); perr == nil && id > last {
			out = append(out, e)
		}
	}
	return out
}

// writeEvent renders one SSE frame (id / event / data lines).
func writeEvent(w io.Writer, e Event) {
	if e.ID != "" {
		fmt.Fprintf(w, "id: %s\n", e.ID)
	}
	if e.Type != "" {
		fmt.Fprintf(w, "event: %s\n", e.Type)
	}
	for _, line := range bytes.Split(e.Data, []byte("\n")) {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprint(w, "\n")
}

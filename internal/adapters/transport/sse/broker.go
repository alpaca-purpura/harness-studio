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

// subBuffer is the per-subscriber send buffer. A subscriber that falls this far behind is
// shed (see Publish) rather than dropping frames silently — it reconnects and replays the
// gap via Last-Event-ID (boundary sesion-viva-consistente `sin-perdida-silenciosa`).
const subBuffer = 256

// Event is one SSE message. ID is a monotonic sequence; Type is one of the Event*
// constants; Data is the payload (may be multi-line).
type Event struct {
	ID   string
	Type string
	Data []byte
}

// subscriber is one connected client. The data channel is NEVER closed (so Publish can
// never send on a closed channel); disconnection is signalled by closing done, which both
// the ServeHTTP reader and Publish observe. sync.Once makes disconnect idempotent across
// the two callers (a lagged shed in Publish vs. the ServeHTTP defer).
type subscriber struct {
	ch   chan Event
	done chan struct{}
	once sync.Once
}

func (s *subscriber) disconnect() { s.once.Do(func() { close(s.done) }) }

// Broker fans events out to every connected subscriber. Safe for concurrent use.
type Broker struct {
	mu      sync.RWMutex
	subs    map[*subscriber]struct{}
	history []Event
	seq     uint64
}

// NewBroker returns an empty broker.
func NewBroker() *Broker {
	return &Broker{subs: map[*subscriber]struct{}{}}
}

// Publish assigns the next id, records the event for replay, and fans it out. It returns
// the stored event (with its assigned id). A subscriber whose buffer is full is not
// silently skipped: it is SHED (disconnected) so its EventSource reconnects and replays the
// gap from history via Last-Event-ID. This turns silent loss into guaranteed catch-up as
// long as the gap stays within maxHistory.
func (b *Broker) Publish(eventType string, data []byte) Event {
	b.mu.Lock()
	b.seq++
	e := Event{ID: strconv.FormatUint(b.seq, 10), Type: eventType, Data: data}
	b.history = append(b.history, e)
	if len(b.history) > maxHistory {
		b.history = b.history[len(b.history)-maxHistory:]
	}
	subs := make([]*subscriber, 0, len(b.subs))
	for s := range b.subs {
		subs = append(subs, s)
	}
	b.mu.Unlock()

	var lagged []*subscriber
	for _, s := range subs {
		select {
		case s.ch <- e:
		case <-s.done: // already disconnected — skip.
		default: // buffer full: shed this subscriber, it catches up on reconnect.
			lagged = append(lagged, s)
		}
	}
	if len(lagged) > 0 {
		b.mu.Lock()
		for _, s := range lagged {
			delete(b.subs, s)
			s.disconnect()
		}
		b.mu.Unlock()
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
	_, _ = fmt.Fprint(w, ": connected\n\n") // open the stream; a dead client surfaces as ctx.Done below.
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-sub.done: // shed as a lagging subscriber — client reconnects + replays.
			return
		case e := <-sub.ch:
			writeEvent(w, e)
			flusher.Flush()
		}
	}
}

func (b *Broker) subscribe() *subscriber {
	s := &subscriber{ch: make(chan Event, subBuffer), done: make(chan struct{})}
	b.mu.Lock()
	b.subs[s] = struct{}{}
	b.mu.Unlock()
	return s
}

func (b *Broker) unsubscribe(s *subscriber) {
	b.mu.Lock()
	delete(b.subs, s)
	b.mu.Unlock()
	s.disconnect()
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

// writeEvent renders one SSE frame (id / event / data lines). Write errors are
// deliberately discarded: SSE has no in-band error channel, and a broken connection
// surfaces as the request context's Done in ServeHTTP's loop — the client then
// reconnects and replays the gap via Last-Event-ID (never a silent loss).
func writeEvent(w io.Writer, e Event) {
	if e.ID != "" {
		_, _ = fmt.Fprintf(w, "id: %s\n", e.ID)
	}
	if e.Type != "" {
		_, _ = fmt.Fprintf(w, "event: %s\n", e.Type)
	}
	for _, line := range bytes.Split(e.Data, []byte("\n")) {
		_, _ = fmt.Fprintf(w, "data: %s\n", line)
	}
	_, _ = fmt.Fprint(w, "\n")
}

package audit

import (
	"encoding/json"
	"sync"
	"time"
)

// Event represents an audit event.
type Event struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// NewEvent creates an audit event from metric names and client IP.
func NewEvent(metricNames []string, ipAddress string) Event {
	return Event{
		TS:        time.Now().Unix(),
		Metrics:   metricNames,
		IPAddress: ipAddress,
	}
}

// MarshalJSON returns the JSON encoding of the event.
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal((Alias)(e))
}

// Listener is the Observer interface for audit events.
type Listener interface {
	OnEvent(event Event) error
}

// ListenerFunc is an adapter to use ordinary functions as Listener.
type ListenerFunc func(Event) error

func (f ListenerFunc) OnEvent(event Event) error {
	return f(event)
}

// Publisher manages audit listeners and notifies them about events.
type Publisher struct {
	mu        sync.RWMutex
	listeners []Listener
}

// NewPublisher creates a new Publisher.
func NewPublisher() *Publisher {
	return &Publisher{}
}

// Subscribe adds a listener. This method is safe for concurrent use.
func (p *Publisher) Subscribe(l Listener) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.listeners = append(p.listeners, l)
}

// Unsubscribe removes a listener. This method is safe for concurrent use.
func (p *Publisher) Unsubscribe(l Listener) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, listener := range p.listeners {
		if listener == l {
			p.listeners = append(p.listeners[:i], p.listeners[i+1:]...)
			return
		}
	}
}

// Publish notifies all listeners about the event.
// Errors from individual listeners are logged but do not stop notification of others.
func (p *Publisher) Publish(event Event) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, l := range p.listeners {
		_ = l.OnEvent(event)
	}
}

// HasListeners returns true if there are any registered listeners.
func (p *Publisher) HasListeners() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.listeners) > 0
}

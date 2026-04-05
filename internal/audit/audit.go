package audit

import (
	"encoding/json"
	"time"
)

// Event represents an audit event.
type Event struct {
	Ts        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// NewEvent creates an audit event from metric names and client IP.
func NewEvent(metricNames []string, ipAddress string) Event {
	return Event{
		Ts:        time.Now().Unix(),
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
	listeners []Listener
}

// NewPublisher creates a new Publisher.
func NewPublisher() *Publisher {
	return &Publisher{}
}

// Subscribe adds a listener.
func (p *Publisher) Subscribe(l Listener) {
	p.listeners = append(p.listeners, l)
}

// Publish notifies all listeners about the event.
// Errors from individual listeners are logged but do not stop notification of others.
func (p *Publisher) Publish(event Event) {
	for _, l := range p.listeners {
		_ = l.OnEvent(event)
	}
}

// HasListeners returns true if there are any registered listeners.
func (p *Publisher) HasListeners() bool {
	return len(p.listeners) > 0
}

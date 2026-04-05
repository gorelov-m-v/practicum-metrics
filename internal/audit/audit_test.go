package audit

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewEvent(t *testing.T) {
	names := []string{"Alloc", "Frees"}
	event := NewEvent(names, "192.168.0.42")

	if event.TS == 0 {
		t.Error("expected non-zero timestamp")
	}
	if len(event.Metrics) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(event.Metrics))
	}
	if event.IPAddress != "192.168.0.42" {
		t.Errorf("expected IP 192.168.0.42, got %s", event.IPAddress)
	}
}

func TestPublisher(t *testing.T) {
	pub := NewPublisher()
	if pub.HasListeners() {
		t.Error("expected no listeners")
	}

	var received []Event
	pub.Subscribe(ListenerFunc(func(e Event) error {
		received = append(received, e)
		return nil
	}))

	if !pub.HasListeners() {
		t.Error("expected listeners after subscribe")
	}

	event := NewEvent([]string{"Alloc"}, "127.0.0.1")
	pub.Publish(event)

	if len(received) != 1 {
		t.Fatalf("expected 1 event, got %d", len(received))
	}
	if received[0].IPAddress != "127.0.0.1" {
		t.Errorf("expected IP 127.0.0.1, got %s", received[0].IPAddress)
	}
}

func TestFileListener(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	listener := NewFileListener(path)
	event := NewEvent([]string{"Alloc", "Frees"}, "10.0.0.1")

	if err := listener.OnEvent(event); err != nil {
		t.Fatalf("OnEvent failed: %v", err)
	}
	if err := listener.OnEvent(event); err != nil {
		t.Fatalf("OnEvent failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	var parsed Event
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if parsed.IPAddress != "10.0.0.1" {
		t.Errorf("expected IP 10.0.0.1, got %s", parsed.IPAddress)
	}
}

func TestURLListener(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	listener := NewURLListener(server.URL)
	event := NewEvent([]string{"HeapAlloc"}, "172.16.0.1")

	if err := listener.OnEvent(event); err != nil {
		t.Fatalf("OnEvent failed: %v", err)
	}

	var parsed Event
	if err := json.Unmarshal(receivedBody, &parsed); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if parsed.IPAddress != "172.16.0.1" {
		t.Errorf("expected IP 172.16.0.1, got %s", parsed.IPAddress)
	}
	if len(parsed.Metrics) != 1 || parsed.Metrics[0] != "HeapAlloc" {
		t.Errorf("unexpected metrics: %v", parsed.Metrics)
	}
}

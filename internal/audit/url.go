package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// URLListener sends audit events via HTTP POST to a remote server.
type URLListener struct {
	url    string
	client *http.Client
}

// NewURLListener creates a new URLListener for the given URL.
func NewURLListener(url string) *URLListener {
	return &URLListener{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// OnEvent sends the event as JSON via POST to the configured URL.
func (u *URLListener) OnEvent(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	resp, err := u.client.Post(u.url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("audit URL returned status %d", resp.StatusCode)
	}
	return nil
}

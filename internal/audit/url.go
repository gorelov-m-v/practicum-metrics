package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// URLListener sends audit events via HTTP POST to a remote server.
type URLListener struct {
	url    string
	client *retryablehttp.Client
}

// NewURLListener creates a new URLListener for the given URL.
func NewURLListener(url string) *URLListener {
	client := retryablehttp.NewClient()
	client.HTTPClient.Timeout = 5 * time.Second
	client.RetryMax = 3
	client.Logger = nil
	return &URLListener{
		url:    url,
		client: client,
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

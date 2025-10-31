package agent

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	maxRetries   = 3
	retryDelay   = 1 * time.Second
	retryBackoff = 2 * time.Second
)

type MetricsSender struct {
	serverAddress string
	client        HTTPClient
}

func NewMetricsSender(serverAddress string) *MetricsSender {
	restyClient := resty.New().
		SetTimeout(5*time.Second).
		SetHeader("Content-Type", "text/plain")

	return NewMetricsSenderWithClient(serverAddress, NewRestyClientAdapter(restyClient))
}

func NewMetricsSenderWithClient(serverAddress string, client HTTPClient) *MetricsSender {
	return &MetricsSender{
		serverAddress: serverAddress,
		client:        client,
	}
}

func (ms *MetricsSender) SendGauge(name string, value float64) error {
	reqURL, err := url.JoinPath(ms.serverAddress, "update", "gauge", url.PathEscape(name), fmt.Sprintf("%v", value))
	if err != nil {
		return fmt.Errorf("failed to build URL: %w", err)
	}

	var lastErr error
	delay := retryDelay

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := ms.client.Post(reqURL)
		if err == nil && resp.StatusCode() == http.StatusOK {
			return nil
		}

		if err != nil {
			lastErr = fmt.Errorf("failed to send metric: %w", err)
		} else {
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		if attempt < maxRetries-1 {
			time.Sleep(delay)
			delay += retryBackoff
		}
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func (ms *MetricsSender) SendCounter(name string, value int64) error {
	reqURL, err := url.JoinPath(ms.serverAddress, "update", "counter", url.PathEscape(name), fmt.Sprintf("%d", value))
	if err != nil {
		return fmt.Errorf("failed to build URL: %w", err)
	}

	var lastErr error
	delay := retryDelay

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := ms.client.Post(reqURL)
		if err == nil && resp.StatusCode() == http.StatusOK {
			return nil
		}

		if err != nil {
			lastErr = fmt.Errorf("failed to send metric: %w", err)
		} else {
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		if attempt < maxRetries-1 {
			time.Sleep(delay)
			delay += retryBackoff
		}
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

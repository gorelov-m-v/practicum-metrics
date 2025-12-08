package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/storage"
)

const (
	maxRetries     = 3
	retryDelay     = 1 * time.Second
	retryBackoff   = 2 * time.Second
	defaultTimeout = 5 * time.Second
)

type MetricsSender struct {
	serverAddress string
	client        HTTPClient
	httpClient    *http.Client
}

func NewMetricsSender(serverAddress string) *MetricsSender {
	restyClient := resty.New().
		SetTimeout(defaultTimeout).
		SetHeader("Content-Type", "text/plain")

	return &MetricsSender{
		serverAddress: serverAddress,
		client:        NewRestyClientAdapter(restyClient),
		httpClient:    &http.Client{Timeout: defaultTimeout},
	}
}

func NewMetricsSenderWithClient(serverAddress string, client HTTPClient) *MetricsSender {
	return &MetricsSender{
		serverAddress: serverAddress,
		client:        client,
		httpClient:    &http.Client{Timeout: defaultTimeout},
	}
}

func (ms *MetricsSender) sendMetric(metricType, name, value string) error {
	reqURL, err := url.JoinPath(ms.serverAddress, "update", metricType, url.PathEscape(name), value)
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

func (ms *MetricsSender) SendGauge(name string, value float64) error {
	return ms.sendMetric(string(storage.Gauge), name, fmt.Sprintf("%v", value))
}

func (ms *MetricsSender) SendCounter(name string, value int64) error {
	return ms.sendMetric(string(storage.Counter), name, fmt.Sprintf("%d", value))
}

func (ms *MetricsSender) sendMetricJSON(metric model.Metrics) error {
	reqURL, err := url.JoinPath(ms.serverAddress, "update")
	if err != nil {
		return fmt.Errorf("failed to build URL: %w", err)
	}

	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	if _, err := gzWriter.Write(body); err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	var lastErr error
	delay := retryDelay

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(buf.Bytes()))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := ms.httpClient.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		} else {
			lastErr = fmt.Errorf("failed to send metric: %w", err)
		}

		if attempt < maxRetries-1 {
			time.Sleep(delay)
			delay += retryBackoff
		}
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func (ms *MetricsSender) SendGaugeJSON(name string, value float64) error {
	metric := model.Metrics{
		ID:    name,
		MType: string(storage.Gauge),
		Value: &value,
	}
	return ms.sendMetricJSON(metric)
}

func (ms *MetricsSender) SendCounterJSON(name string, value int64) error {
	metric := model.Metrics{
		ID:    name,
		MType: string(storage.Counter),
		Delta: &value,
	}
	return ms.sendMetricJSON(metric)
}

func (ms *MetricsSender) SendMetricsBatch(metrics []model.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	reqURL, err := url.JoinPath(ms.serverAddress, "updates")
	if err != nil {
		return fmt.Errorf("failed to build URL: %w", err)
	}

	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	if _, err := gzWriter.Write(body); err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	var lastErr error
	delay := retryDelay

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(buf.Bytes()))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := ms.httpClient.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		} else {
			lastErr = fmt.Errorf("failed to send metrics batch: %w", err)
		}

		if attempt < maxRetries-1 {
			time.Sleep(delay)
			delay += retryBackoff
		}
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

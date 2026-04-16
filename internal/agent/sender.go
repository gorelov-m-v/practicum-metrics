package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/user/practicum-metrics/internal/encryption"
	"github.com/user/practicum-metrics/internal/hash"
	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/retry"
	"github.com/user/practicum-metrics/internal/storage"
)

const (
	defaultTimeout = 5 * time.Second
)

type MetricsSender struct {
	serverAddress string
	client        HTTPClient
	httpClient    *http.Client
	key           string
	publicKey     *rsa.PublicKey
}

func NewMetricsSender(serverAddress string, key string) *MetricsSender {
	restyClient := resty.New().
		SetTimeout(defaultTimeout).
		SetHeader("Content-Type", "text/plain")

	return &MetricsSender{
		serverAddress: serverAddress,
		client:        NewRestyClientAdapter(restyClient),
		httpClient:    &http.Client{Timeout: defaultTimeout},
		key:           key,
	}
}

func NewMetricsSenderWithClient(serverAddress string, client HTTPClient) *MetricsSender {
	return &MetricsSender{
		serverAddress: serverAddress,
		client:        client,
		httpClient:    &http.Client{Timeout: defaultTimeout},
		key:           "",
	}
}

func NewMetricsSenderWithClientAndKey(serverAddress string, client HTTPClient, key string) *MetricsSender {
	return &MetricsSender{
		serverAddress: serverAddress,
		client:        client,
		httpClient:    &http.Client{Timeout: defaultTimeout},
		key:           key,
	}
}

func (ms *MetricsSender) SetPublicKey(publicKey *rsa.PublicKey) {
	ms.publicKey = publicKey
}

func (ms *MetricsSender) encodeRequestBody(body []byte) ([]byte, bool, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	if _, err := gzWriter.Write(body); err != nil {
		return nil, false, fmt.Errorf("failed to compress data: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return nil, false, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	if ms.publicKey == nil {
		return buf.Bytes(), false, nil
	}

	encryptedBody, err := encryption.Encrypt(buf.Bytes(), ms.publicKey)
	if err != nil {
		return nil, false, fmt.Errorf("failed to encrypt body: %w", err)
	}

	return encryptedBody, true, nil
}

func (ms *MetricsSender) sendMetric(metricType, name, value string) error {
	reqURL, err := url.JoinPath(ms.serverAddress, "update", metricType, url.PathEscape(name), value)
	if err != nil {
		return fmt.Errorf("failed to build URL: %w", err)
	}

	return retry.Do(func() error {
		resp, err := ms.client.Post(reqURL)
		if err != nil {
			return fmt.Errorf("failed to send metric: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}
		return nil
	})
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

	payload, encrypted, err := ms.encodeRequestBody(body)
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(payload))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		if encrypted {
			req.Header.Set(encryption.HeaderEncrypted, encryption.HeaderEncryptedValue)
		}

		if ms.key != "" {
			hashValue := hash.CalculateHMAC(body, ms.key)
			req.Header.Set("HashSHA256", hashValue)
		}

		resp, err := ms.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send metric: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return nil
	})
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

	payload, encrypted, err := ms.encodeRequestBody(body)
	if err != nil {
		return err
	}

	return retry.Do(func() error {
		req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(payload))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		if encrypted {
			req.Header.Set(encryption.HeaderEncrypted, encryption.HeaderEncryptedValue)
		}

		if ms.key != "" {
			hashValue := hash.CalculateHMAC(body, ms.key)
			req.Header.Set("HashSHA256", hashValue)
		}

		resp, err := ms.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send metrics batch: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return nil
	})
}

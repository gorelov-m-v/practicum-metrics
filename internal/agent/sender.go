package agent

import (
	"fmt"
	"net/http"
	"time"
)

type MetricsSender struct {
	serverAddress string
	client        *http.Client
}

func NewMetricsSender(serverAddress string) *MetricsSender {
	return &MetricsSender{
		serverAddress: serverAddress,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (ms *MetricsSender) SendGauge(name string, value float64) error {
	url := fmt.Sprintf("%s/update/gauge/%s/%v", ms.serverAddress, name, value)
	return ms.doRequest(url)
}

func (ms *MetricsSender) SendCounter(name string, value int64) error {
	url := fmt.Sprintf("%s/update/counter/%s/%d", ms.serverAddress, name, value)
	return ms.doRequest(url)
}

func (ms *MetricsSender) doRequest(url string) error {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := ms.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metric: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

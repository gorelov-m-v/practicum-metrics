package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type MetricsSender struct {
	serverAddress string
	client        *resty.Client
}

func NewMetricsSender(serverAddress string) *MetricsSender {
	return &MetricsSender{
		serverAddress: serverAddress,
		client: resty.New().
			SetTimeout(5*time.Second).
			SetHeader("Content-Type", "text/plain"),
	}
}

func (ms *MetricsSender) SendGauge(name string, value float64) error {
	url := fmt.Sprintf("%s/update/gauge/%s/%v", ms.serverAddress, name, value)
	resp, err := ms.client.R().Post(url)
	if err != nil {
		return fmt.Errorf("failed to send metric: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	return nil
}

func (ms *MetricsSender) SendCounter(name string, value int64) error {
	url := fmt.Sprintf("%s/update/counter/%s/%d", ms.serverAddress, name, value)
	resp, err := ms.client.R().Post(url)
	if err != nil {
		return fmt.Errorf("failed to send metric: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	return nil
}

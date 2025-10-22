package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/user/practicum-metrics/internal/storage"
)

type MetricHandler struct {
	storage storage.Storage
}

func NewMetricHandler(s storage.Storage) *MetricHandler {
	return &MetricHandler{
		storage: s,
	}
}

func (h *MetricHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/update/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || parts[1] == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Metric value is required", http.StatusBadRequest)
		return
	}

	metricType := parts[0]
	metricName := parts[1]
	metricValue := parts[2]

	switch storage.MetricType(metricType) {
	case storage.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(metricName, value)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

	case storage.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(metricName, value)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
	}
}

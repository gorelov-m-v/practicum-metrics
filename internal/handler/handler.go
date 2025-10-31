package handler

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
)

//go:embed templates/metrics.html
var metricsTemplate string

type MetricHandler struct {
	service *service.MetricsService
}

type metricData struct {
	Type  string
	Name  string
	Value string
}

type metricsPageData struct {
	Metrics []metricData
}

func NewMetricHandler(s *service.MetricsService) *MetricHandler {
	return &MetricHandler{
		service: s,
	}
}

func (h *MetricHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	switch storage.MetricType(metricType) {
	case storage.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateGauge(metricName, value); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

	case storage.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateCounter(metricName, value); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
	}
}

func (h *MetricHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	switch storage.MetricType(metricType) {
	case storage.Gauge:
		value, exists := h.service.GetGauge(metricName)
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%g", value)

	case storage.Counter:
		value, exists := h.service.GetCounter(metricName)
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%d", value)

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
	}
}

func (h *MetricHandler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	gauges := h.service.GetAllGauges()
	counters := h.service.GetAllCounters()

	var metrics []metricData

	for name, value := range gauges {
		metrics = append(metrics, metricData{
			Type:  "gauge",
			Name:  name,
			Value: fmt.Sprintf("%g", value),
		})
	}

	for name, value := range counters {
		metrics = append(metrics, metricData{
			Type:  "counter",
			Name:  name,
			Value: fmt.Sprintf("%d", value),
		})
	}

	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].Type != metrics[j].Type {
			return metrics[i].Type < metrics[j].Type
		}
		return metrics[i].Name < metrics[j].Name
	})

	tmpl, err := template.New("metrics").Parse(metricsTemplate)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := metricsPageData{Metrics: metrics}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

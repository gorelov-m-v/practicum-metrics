package handler

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
)

//go:embed all:templates/metrics.html
var metricsTemplate string

type MetricHandler struct {
	service   *service.MetricsService
	persister *storage.Persister
	template  *template.Template
}

type metricData struct {
	Type  string
	Name  string
	Value string
}

type metricsPageData struct {
	Metrics []metricData
}

func NewMetricHandler(s *service.MetricsService, p *storage.Persister) (*MetricHandler, error) {
	tmpl, err := template.New("metrics").Parse(metricsTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded template: %w", err)
	}

	return &MetricHandler{
		service:   s,
		persister: p,
		template:  tmpl,
	}, nil
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
		if h.persister != nil && h.persister.IsSyncMode() {
			h.persister.SaveSync()
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
		if h.persister != nil && h.persister.IsSyncMode() {
			h.persister.SaveSync()
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
			Type:  string(storage.Gauge),
			Name:  name,
			Value: fmt.Sprintf("%g", value),
		})
	}

	for name, value := range counters {
		metrics = append(metrics, metricData{
			Type:  string(storage.Counter),
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := metricsPageData{Metrics: metrics}
	if err := h.template.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *MetricHandler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	var req model.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp := model.Metrics{
		ID:    req.ID,
		MType: req.MType,
	}

	switch storage.MetricType(req.MType) {
	case storage.Gauge:
		if req.Value == nil {
			http.Error(w, "Missing value for gauge", http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateGauge(req.ID, *req.Value); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if h.persister != nil && h.persister.IsSyncMode() {
			h.persister.SaveSync()
		}
		value, _ := h.service.GetGauge(req.ID)
		resp.Value = &value

	case storage.Counter:
		if req.Delta == nil {
			http.Error(w, "Missing delta for counter", http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateCounter(req.ID, *req.Delta); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if h.persister != nil && h.persister.IsSyncMode() {
			h.persister.SaveSync()
		}
		delta, _ := h.service.GetCounter(req.ID)
		resp.Delta = &delta

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		return
	}
}

func (h *MetricHandler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	var req model.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp := model.Metrics{
		ID:    req.ID,
		MType: req.MType,
	}

	switch storage.MetricType(req.MType) {
	case storage.Gauge:
		value, exists := h.service.GetGauge(req.ID)
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		resp.Value = &value

	case storage.Counter:
		value, exists := h.service.GetCounter(req.ID)
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		resp.Delta = &value

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		return
	}
}

package handler

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/user/practicum-metrics/internal/audit"
	"github.com/user/practicum-metrics/internal/database"
	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
)

//go:embed all:templates/metrics.html
var metricsTemplate string

type MetricHandler struct {
	service        *service.MetricsService
	persister      *storage.Persister
	template       *template.Template
	db             *database.DB
	auditPublisher *audit.Publisher
}

type metricData struct {
	Type  string
	Name  string
	Value string
}

type metricsPageData struct {
	Metrics []metricData
}

func NewMetricHandler(s *service.MetricsService, p *storage.Persister, db *database.DB, ap *audit.Publisher) (*MetricHandler, error) {
	tmpl, err := template.New("metrics").Parse(metricsTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded template: %w", err)
	}

	return &MetricHandler{
		service:        s,
		persister:      p,
		template:       tmpl,
		db:             db,
		auditPublisher: ap,
	}, nil
}

func (h *MetricHandler) publishAudit(r *http.Request, metricNames []string) {
	if h.auditPublisher == nil || !h.auditPublisher.HasListeners() {
		return
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	h.auditPublisher.Publish(audit.NewEvent(metricNames, ip))
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
		if err := h.service.UpdateGauge(r.Context(), metricName, value); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if h.persister != nil && h.persister.IsSyncMode() {
			h.persister.SaveSync()
		}
		h.publishAudit(r, []string{metricName})
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

	case storage.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateCounter(r.Context(), metricName, value); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if h.persister != nil && h.persister.IsSyncMode() {
			h.persister.SaveSync()
		}
		h.publishAudit(r, []string{metricName})
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
		value, exists := h.service.GetGauge(r.Context(), metricName)
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%g", value)

	case storage.Counter:
		value, exists := h.service.GetCounter(r.Context(), metricName)
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
	gauges := h.service.GetAllGauges(r.Context())
	counters := h.service.GetAllCounters(r.Context())

	metrics := make([]metricData, 0, len(gauges)+len(counters))

	gaugeType := string(storage.Gauge)
	counterType := string(storage.Counter)

	for name, value := range gauges {
		metrics = append(metrics, metricData{
			Type:  gaugeType,
			Name:  name,
			Value: strconv.FormatFloat(value, 'g', -1, 64),
		})
	}

	for name, value := range counters {
		metrics = append(metrics, metricData{
			Type:  counterType,
			Name:  name,
			Value: strconv.FormatInt(value, 10),
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
		if err := h.service.UpdateGauge(r.Context(), req.ID, *req.Value); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if h.persister != nil && h.persister.IsSyncMode() {
			h.persister.SaveSync()
		}
		h.publishAudit(r, []string{req.ID})
		value, _ := h.service.GetGauge(r.Context(), req.ID)
		resp.Value = &value

	case storage.Counter:
		if req.Delta == nil {
			http.Error(w, "Missing delta for counter", http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateCounter(r.Context(), req.ID, *req.Delta); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if h.persister != nil && h.persister.IsSyncMode() {
			h.persister.SaveSync()
		}
		h.publishAudit(r, []string{req.ID})
		delta, _ := h.service.GetCounter(r.Context(), req.ID)
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
		value, exists := h.service.GetGauge(r.Context(), req.ID)
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		resp.Value = &value

	case storage.Counter:
		value, exists := h.service.GetCounter(r.Context(), req.ID)
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

func (h *MetricHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	// PingDB is infrastructure, not business logic - timeout stays here
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		http.Error(w, "Database ping failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) UpdateMetricsBatch(w http.ResponseWriter, r *http.Request) {
	var metrics []model.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(metrics) == 0 {
		http.Error(w, "Empty batch not allowed", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateMetricsBatch(r.Context(), metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.persister != nil && h.persister.IsSyncMode() {
		h.persister.SaveSync()
	}

	names := make([]string, len(metrics))
	for i, m := range metrics {
		names[i] = m.ID
	}
	h.publishAudit(r, names)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

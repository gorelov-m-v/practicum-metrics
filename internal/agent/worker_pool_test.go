package agent

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/user/practicum-metrics/internal/model"
)

func TestNewWorkerPool(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sender := NewMetricsSender("http://localhost:8080", "")

	tests := []struct {
		name    string
		workers int
	}{
		{"single worker", 1},
		{"multiple workers", 5},
		{"many workers", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewWorkerPool(tt.workers, sender, logger)
			if pool == nil {
				t.Fatal("NewWorkerPool returned nil")
			}
			if pool.workers != tt.workers {
				t.Errorf("expected %d workers, got %d", tt.workers, pool.workers)
			}
			if pool.taskQueue == nil {
				t.Error("taskQueue not initialized")
			}
		})
	}
}

func TestWorkerPool_StartStop(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sender := NewMetricsSender("http://localhost:8080", "")

	tests := []struct {
		name    string
		workers int
	}{
		{"single worker", 1},
		{"multiple workers", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewWorkerPool(tt.workers, sender, logger)
			pool.Start()

			time.Sleep(10 * time.Millisecond)

			pool.Stop()

			select {
			case <-pool.ctx.Done():
			default:
				t.Error("context should be cancelled after Stop()")
			}
		})
	}
}

func TestWorkerPool_Submit(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewMetricsSender(server.URL, "")

	tests := []struct {
		name           string
		workers        int
		tasksCount     int
		metricsPerTask int
	}{
		{"single task single worker", 1, 1, 5},
		{"multiple tasks single worker", 1, 5, 3},
		{"multiple tasks multiple workers", 3, 10, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atomic.StoreInt32(&requestCount, 0)

			pool := NewWorkerPool(tt.workers, sender, logger)
			pool.Start()

			for i := 0; i < tt.tasksCount; i++ {
				var metrics []model.Metrics
				for j := 0; j < tt.metricsPerTask; j++ {
					value := float64(j)
					metrics = append(metrics, model.Metrics{
						ID:    "test_metric",
						MType: "gauge",
						Value: &value,
					})
				}
				pool.Submit(MetricTask{Metrics: metrics})
			}

			time.Sleep(100 * time.Millisecond)
			pool.Stop()

			expectedRequests := int32(tt.tasksCount)
			actualRequests := atomic.LoadInt32(&requestCount)

			if actualRequests != expectedRequests {
				t.Errorf("expected %d requests, got %d", expectedRequests, actualRequests)
			}
		})
	}
}

func TestWorkerPool_RateLimiting(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	var mu sync.Mutex
	var concurrentRequests int32
	var maxConcurrent int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&concurrentRequests, 1)

		mu.Lock()
		if current > maxConcurrent {
			maxConcurrent = current
		}
		mu.Unlock()

		time.Sleep(50 * time.Millisecond)

		atomic.AddInt32(&concurrentRequests, -1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewMetricsSender(server.URL, "")
	maxWorkers := 3

	pool := NewWorkerPool(maxWorkers, sender, logger)
	pool.Start()

	tasksCount := 20
	for i := 0; i < tasksCount; i++ {
		value := float64(i)
		metrics := []model.Metrics{
			{
				ID:    "test_metric",
				MType: "gauge",
				Value: &value,
			},
		}
		pool.Submit(MetricTask{Metrics: metrics})
	}

	pool.Stop()

	mu.Lock()
	finalMaxConcurrent := maxConcurrent
	mu.Unlock()

	if finalMaxConcurrent > int32(maxWorkers) {
		t.Errorf("rate limiting failed: max concurrent requests was %d, but limit is %d", finalMaxConcurrent, maxWorkers)
	}

	t.Logf("Max concurrent requests: %d (limit: %d)", finalMaxConcurrent, maxWorkers)
}

func TestWorkerPool_SubmitAfterStop(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sender := NewMetricsSender("http://localhost:8080", "")

	pool := NewWorkerPool(2, sender, logger)
	pool.Start()
	pool.Stop()

	value := 1.0
	metrics := []model.Metrics{
		{
			ID:    "test_metric",
			MType: "gauge",
			Value: &value,
		},
	}

	pool.Submit(MetricTask{Metrics: metrics})
}

func TestWorkerPool_ConcurrentSubmit(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewMetricsSender(server.URL, "")

	pool := NewWorkerPool(5, sender, logger)
	pool.Start()

	goroutines := 10
	tasksPerGoroutine := 5
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < tasksPerGoroutine; j++ {
				value := float64(j)
				metrics := []model.Metrics{
					{
						ID:    "test_metric",
						MType: "gauge",
						Value: &value,
					},
				}
				pool.Submit(MetricTask{Metrics: metrics})
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	pool.Stop()

	expectedRequests := int32(goroutines * tasksPerGoroutine)
	actualRequests := atomic.LoadInt32(&requestCount)

	if actualRequests != expectedRequests {
		t.Errorf("expected %d requests, got %d", expectedRequests, actualRequests)
	}
}

func TestWorkerPool_ErrorHandling(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := NewMetricsSender(server.URL, "")

	pool := NewWorkerPool(2, sender, logger)
	pool.Start()

	tasksCount := 3
	for i := 0; i < tasksCount; i++ {
		value := float64(i)
		metrics := []model.Metrics{
			{
				ID:    "test_metric",
				MType: "gauge",
				Value: &value,
			},
		}
		pool.Submit(MetricTask{Metrics: metrics})
	}

	pool.Stop()

	actualRequests := atomic.LoadInt32(&requestCount)
	if actualRequests == 0 {
		t.Error("no requests were made despite errors")
	}

	t.Logf("Requests made despite errors: %d", actualRequests)
}

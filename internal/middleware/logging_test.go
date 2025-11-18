package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogging(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		handler        http.HandlerFunc
		expectedStatus int
		expectedSize   int
	}{
		{
			name:   "GET request with 200 OK",
			method: "GET",
			path:   "/test",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello, World!"))
			},
			expectedStatus: http.StatusOK,
			expectedSize:   13,
		},
		{
			name:   "POST request with 201 Created",
			method: "POST",
			path:   "/create",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("Created"))
			},
			expectedStatus: http.StatusCreated,
			expectedSize:   7,
		},
		{
			name:   "request with 404 Not Found",
			method: "GET",
			path:   "/notfound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Not Found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedSize:   9,
		},
		{
			name:   "request without explicit WriteHeader",
			method: "GET",
			path:   "/implicit",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("OK"))
			},
			expectedStatus: http.StatusOK,
			expectedSize:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, recorded := observer.New(zap.InfoLevel)
			logger := zap.New(core)

			middleware := Logging(logger)
			handler := middleware(tt.handler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			logs := recorded.All()
			if len(logs) != 1 {
				t.Fatalf("expected 1 log entry, got %d", len(logs))
			}

			logEntry := logs[0]

			if logEntry.Level != zap.InfoLevel {
				t.Errorf("expected log level Info, got %v", logEntry.Level)
			}

			if logEntry.Message != "HTTP request" {
				t.Errorf("expected message 'HTTP request', got '%s'", logEntry.Message)
			}

			fields := logEntry.ContextMap()

			if fields["method"] != tt.method {
				t.Errorf("expected method '%s', got '%v'", tt.method, fields["method"])
			}

			if fields["uri"] != tt.path {
				t.Errorf("expected uri '%s', got '%v'", tt.path, fields["uri"])
			}

			if fields["status"] != int64(tt.expectedStatus) {
				t.Errorf("expected status %d, got %v", tt.expectedStatus, fields["status"])
			}

			if fields["size"] != int64(tt.expectedSize) {
				t.Errorf("expected size %d, got %v", tt.expectedSize, fields["size"])
			}

			if _, ok := fields["duration"]; !ok {
				t.Error("expected 'duration' field in log")
			}
		})
	}
}

func TestLogging_MultipleWrites(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("First "))
		w.Write([]byte("Second "))
		w.Write([]byte("Third"))
	})

	middleware := Logging(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/multi", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	logs := recorded.All()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logs))
	}

	fields := logs[0].ContextMap()

	expectedSize := int64(18)
	if fields["size"] != expectedSize {
		t.Errorf("expected size %d, got %v", expectedSize, fields["size"])
	}
}

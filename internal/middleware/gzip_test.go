package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipCompress(t *testing.T) {
	tests := []struct {
		name           string
		acceptEncoding string
		contentType    string
		body           string
		expectGzip     bool
	}{
		{
			name:           "compress JSON with gzip support",
			acceptEncoding: "gzip",
			contentType:    "application/json",
			body:           `{"id":"test","value":123}`,
			expectGzip:     true,
		},
		{
			name:           "compress HTML with gzip support",
			acceptEncoding: "gzip",
			contentType:    "text/html",
			body:           "<html><body>Hello</body></html>",
			expectGzip:     true,
		},
		{
			name:           "no compression without Accept-Encoding",
			acceptEncoding: "",
			contentType:    "application/json",
			body:           `{"id":"test","value":123}`,
			expectGzip:     false,
		},
		{
			name:           "no compression for unsupported content type",
			acceptEncoding: "gzip",
			contentType:    "image/png",
			body:           "binary data",
			expectGzip:     false,
		},
		{
			name:           "compress with deflate, gzip in Accept-Encoding",
			acceptEncoding: "deflate, gzip, br",
			contentType:    "application/json",
			body:           `{"test":true}`,
			expectGzip:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.body))
			})

			middleware := GzipCompress(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if tt.expectGzip {
				if rec.Header().Get("Content-Encoding") != "gzip" {
					t.Errorf("expected Content-Encoding: gzip, got: %s", rec.Header().Get("Content-Encoding"))
				}

				gr, err := gzip.NewReader(rec.Body)
				if err != nil {
					t.Fatalf("failed to create gzip reader: %v", err)
				}
				defer gr.Close()

				decompressed, err := io.ReadAll(gr)
				if err != nil {
					t.Fatalf("failed to read decompressed data: %v", err)
				}

				if string(decompressed) != tt.body {
					t.Errorf("expected body '%s', got '%s'", tt.body, string(decompressed))
				}
			} else {
				if rec.Header().Get("Content-Encoding") == "gzip" {
					t.Error("did not expect Content-Encoding: gzip")
				}

				if rec.Body.String() != tt.body {
					t.Errorf("expected body '%s', got '%s'", tt.body, rec.Body.String())
				}
			}
		})
	}
}

func TestGzipDecompress(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		compress        bool
		contentEncoding string
		expectError     bool
	}{
		{
			name:            "decompress gzip request",
			body:            `{"id":"test","value":456}`,
			compress:        true,
			contentEncoding: "gzip",
			expectError:     false,
		},
		{
			name:            "pass through non-gzip request",
			body:            `{"id":"test","value":789}`,
			compress:        false,
			contentEncoding: "",
			expectError:     false,
		},
		{
			name:            "invalid gzip data",
			body:            "not gzip data",
			compress:        false,
			contentEncoding: "gzip",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.compress {
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				gw.Write([]byte(tt.body))
				gw.Close()
				body = &buf
			} else {
				body = bytes.NewBufferString(tt.body)
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				received, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("failed to read body: %v", err)
				}

				if string(received) != tt.body {
					t.Errorf("expected body '%s', got '%s'", tt.body, string(received))
				}

				w.WriteHeader(http.StatusOK)
			})

			middleware := GzipDecompress(handler)

			req := httptest.NewRequest("POST", "/test", body)
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if tt.expectError {
				if rec.Code != http.StatusBadRequest {
					t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
				}
			} else {
				if rec.Code != http.StatusOK {
					t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
				}

				if req.Header.Get("Content-Encoding") != "" {
					t.Error("expected Content-Encoding header to be removed")
				}
			}
		})
	}
}

func TestGzipCompressAndDecompress(t *testing.T) {
	originalBody := `{"id":"integration","type":"gauge","value":42.5}`

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		if string(body) != originalBody {
			t.Errorf("expected request body '%s', got '%s'", originalBody, string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	middleware := GzipCompress(GzipDecompress(handler))

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte(originalBody))
	gw.Close()

	req := httptest.NewRequest("POST", "/test", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding: gzip, got: %s", rec.Header().Get("Content-Encoding"))
	}

	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to read decompressed data: %v", err)
	}

	if string(decompressed) != originalBody {
		t.Errorf("expected response body '%s', got '%s'", originalBody, string(decompressed))
	}
}

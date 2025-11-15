package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer         io.Writer
	wroteHeader    bool
	shouldCompress bool
}

func (w *gzipWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	contentType := w.ResponseWriter.Header().Get("Content-Type")
	if shouldCompress(contentType) {
		w.shouldCompress = true
		gz, err := gzip.NewWriterLevel(w.ResponseWriter, gzip.DefaultCompression)
		if err == nil {
			w.Writer = gz
			w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
			w.ResponseWriter.Header().Del("Content-Length")
		} else {
			w.shouldCompress = false
		}
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	if w.shouldCompress {
		return w.Writer.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func GzipCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gzw := &gzipWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(gzw, r)

		if gzw.shouldCompress && gzw.Writer != nil {
			if gz, ok := gzw.Writer.(*gzip.Writer); ok {
				gz.Close()
			}
		}
	})
}

func GzipDecompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer gz.Close()

		r.Body = gz

		r.Header.Del("Content-Encoding")
		r.Header.Del("Content-Length")

		next.ServeHTTP(w, r)
	})
}

func shouldCompress(contentType string) bool {
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html")
}

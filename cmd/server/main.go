package main

import (
	"net/http"

	"github.com/user/practicum-metrics/internal/handler"
	"github.com/user/practicum-metrics/internal/storage"
)

func main() {
	store := storage.NewMemStorage()
	h := handler.NewMetricHandler(store)

	err := http.ListenAndServe(`:8080`, http.HandlerFunc(h.UpdateMetric))
	if err != nil {
		panic(err)
	}
}

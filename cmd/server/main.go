package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/user/practicum-metrics/internal/handler"
	"github.com/user/practicum-metrics/internal/storage"
)

func main() {
	parseFlags()

	store := storage.NewMemStorage()
	h := handler.NewMetricHandler(store)

	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", h.UpdateMetric)
	r.Get("/value/{type}/{name}", h.GetMetric)
	r.Get("/", h.ListMetrics)

	if err := http.ListenAndServe(flagRunAddr, r); err != nil {
		panic(err)
	}
}

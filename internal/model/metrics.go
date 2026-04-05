// Package model defines data structures used across the application.
package model

// Metrics represents a single metric value sent between agent and server.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

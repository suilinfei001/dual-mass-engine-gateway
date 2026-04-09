package models

import "time"

// Models export all model types
type Model interface {
	// Common interface for all models if needed
}

// Timestamp helper for models that need UpdatedAt field
type Timestamp struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateTimestamp updates the UpdatedAt field
func (ts *Timestamp) UpdateTimestamp() {
	ts.UpdatedAt = time.Now()
}

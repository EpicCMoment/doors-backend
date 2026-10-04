package dxl

import "encoding/json"

// Envelope is the standard JSON shape every utility script returns.
type Envelope struct {
	Status  string          `json:"status"` // "ok" or "error"
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

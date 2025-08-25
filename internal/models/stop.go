package models

// Stop represents a physical transit stop.
type Stop struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Distance float64 `json:"distance,omitempty"` // Distance in meters
}
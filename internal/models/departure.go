package models

import "time"

type Departure struct {
	Line        string    `json:"line"`
	Destination string    `json:"destination"`
	Departure   time.Time `json:"departure"`
}
package models

import "time"

// Vehicle represents a transit vehicle with real-time location data
type Vehicle struct {
	ID          string    `json:"id"`          // Vehicle identifier
	RouteID     string    `json:"route_id"`    // Route this vehicle is serving
	TripID      string    `json:"trip_id"`     // Current trip identifier
	Lat         float64   `json:"lat"`         // Current latitude
	Lon         float64   `json:"lon"`         // Current longitude
	Bearing     *float64  `json:"bearing,omitempty"` // Direction of travel in degrees (0-359)
	Speed       *float64  `json:"speed,omitempty"`   // Speed in km/h
	Timestamp   time.Time `json:"timestamp"`   // Last update timestamp
	Delay       *int      `json:"delay,omitempty"`   // Delay in seconds (positive = late, negative = early)
	Status      string    `json:"status"`      // Vehicle status (e.g., "IN_TRANSIT", "STOPPED_AT", "INCOMING_AT")
	StopID      *string   `json:"stop_id,omitempty"` // Current or next stop ID
	Occupancy   *string   `json:"occupancy,omitempty"` // Occupancy level ("EMPTY", "MANY_SEATS_AVAILABLE", "FEW_SEATS_AVAILABLE", "STANDING_ROOM_ONLY", "CRUSHED_STANDING_ROOM_ONLY", "FULL", "NOT_ACCEPTING_PASSENGERS")
}

// VehiclePosition represents a simplified vehicle position for tracking
type VehiclePosition struct {
	VehicleID string    `json:"vehicle_id"`
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	Bearing   *float64  `json:"bearing,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// RouteVehicles represents all vehicles currently serving a specific route
type RouteVehicles struct {
	RouteID     string    `json:"route_id"`
	RouteName   string    `json:"route_name"`
	Vehicles    []Vehicle `json:"vehicles"`
	LastUpdated time.Time `json:"last_updated"`
}
package handlers

import (
	"arrivo-transit-api/internal/services"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type TransitHandler struct {
	transitService *services.TransitService
}

func NewTransitHandler(transitService *services.TransitService) *TransitHandler {
	return &TransitHandler{
		transitService: transitService,
	}
}

func (h *TransitHandler) GetDepartures(w http.ResponseWriter, r *http.Request) {
	stopID := chi.URLParam(r, "stopID")
	if stopID == "" {
		http.Error(w, "stopID is required", http.StatusBadRequest)
		return
	}

	departures, err := h.transitService.GetDepartures(r.Context(), stopID)
	if err != nil {
		log.Printf("ERROR: Failed to get departures for stop %s: %v", stopID, err)
		http.Error(w, "Failed to get departures", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(departures)
}

// SearchStops handles searching for stops.
func (h *TransitHandler) SearchStops(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	var lat, lon *float64
	var err error

	if latStr != "" && lonStr != "" {
		latVal, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			http.Error(w, "invalid lat", http.StatusBadRequest)
			return
		}
		lonVal, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			http.Error(w, "invalid lon", http.StatusBadRequest)
			return
		}
		lat = &latVal
		lon = &lonVal
	}

	stops, err := h.transitService.SearchStops(r.Context(), query, lat, lon)
	if err != nil {
		log.Printf("ERROR: Failed to search for stops with query '%s': %v", query, err)
		http.Error(w, "Failed to search for stops", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stops)
}

// GetVehiclesByRoute handles fetching vehicles for a specific route
func (h *TransitHandler) GetVehiclesByRoute(w http.ResponseWriter, r *http.Request) {
	routeID := chi.URLParam(r, "routeID")
	if routeID == "" {
		http.Error(w, "routeID is required", http.StatusBadRequest)
		return
	}

	routeVehicles, err := h.transitService.GetVehiclesByRoute(r.Context(), routeID)
	if err != nil {
		log.Printf("ERROR: Failed to get vehicles for route %s: %v", routeID, err)
		http.Error(w, "Failed to get vehicles for route", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routeVehicles)
}

// GetAllActiveVehicles handles fetching all currently active vehicles
func (h *TransitHandler) GetAllActiveVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.transitService.GetAllActiveVehicles(r.Context())
	if err != nil {
		log.Printf("ERROR: Failed to get all active vehicles: %v", err)
		http.Error(w, "Failed to get active vehicles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vehicles)
}

// SearchRoutes handles route/line search requests
func (h *TransitHandler) SearchRoutes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	routes, err := h.transitService.SearchRoutes(r.Context(), query)
	if err != nil {
		log.Printf("ERROR: Failed to search routes with query '%s': %v", query, err)
		http.Error(w, "Failed to search routes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routes)
}

func (h *TransitHandler) GetNearbyStops(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	radiusStr := r.URL.Query().Get("radius")

	if latStr == "" || lonStr == "" {
		http.Error(w, "lat and lon are required", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "invalid lat", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "invalid lon", http.StatusBadRequest)
		return
	}

	radius := 500.0 // default radius
	if radiusStr != "" {
		radius, err = strconv.ParseFloat(radiusStr, 64)
		if err != nil {
			http.Error(w, "invalid radius", http.StatusBadRequest)
			return
		}
	}

	stops, err := h.transitService.GetNearbyStops(r.Context(), lat, lon, radius)
	if err != nil {
		log.Printf("ERROR: Failed to get nearby stops: %v", err)
		http.Error(w, "Failed to get nearby stops", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stops)
}